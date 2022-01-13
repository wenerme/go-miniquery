package gormq

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func getPreparedDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.sqlite3"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(User{}, UserProfile{}))
	db = db.Debug()
	db.Create(&User{
		Username: "wener",
		FullName: "Wener",
	})
	db.Create(&User{
		Username: "xxx",
		FullName: "XX",
		Profile: &UserProfile{
			Age: 18,
		},
	})
	return db
}

func TestQueryBuild(t *testing.T) {
	db := getPreparedDB(t)
	rows, err := db.Model(User{}).Scopes(ApplyMiniQuery("1=1")).Rows()
	assert.NoError(t, err)
	assert.NotEmpty(t, rows)
	for _, test := range []struct {
		Q     string
		Where string
		Vars  []interface{}
		Err   bool
	}{
		// pg do not support reference alias in where
		// {Q: `Profile.age > 1`, Where: "`Profile__age` > ?", Vars: []interface{}{1}},
		{Q: `Profile.AGE > 1`, Where: "`Profile`.`age` > ?", Vars: []interface{}{1}},
		{Q: `Username = 'wener' and fullName is not null`, Where: "`username` = ? and `full_name` is not null", Vars: []interface{}{"wener"}},
		{Q: `2021 = date(CreatedAt)`, Where: "? = date(`created_at`)", Vars: []interface{}{2021}},
		{Q: `2021 > 0`},
		{Q: `1=2`},
		{Q: `2021 between 1 and 2`},
		{Q: `2021 between 1 and`, Err: true},
	} {
		m := User{}
		query := db.Model(User{}).Scopes(MiniQuery{Query: []string{test.Q}}.Scope).Session(&gorm.Session{DryRun: true})
		assert.NoError(t, query.Error)
		query = query.Find(&m)
		if test.Err {
			assert.Error(t, query.Error)
			continue
		}
		assert.NoError(t, query.Error)
		stat := query.Statement
		s := stat.SQL.String()
		substr := "WHERE "
		fmt.Println("SQL:", s, stat.Vars)
		idx := strings.Index(s, substr)
		w := ""
		if idx > 0 {
			w = s[idx+len(substr):]
		}
		if test.Where != "" {
			assert.Equal(t, test.Where, w)
		}
		if len(test.Vars) != 0 {
			assert.Equal(t, test.Vars, stat.Vars)
		}
		assert.NoError(t, db.Exec(s, stat.Vars...).Error)
	}
}

func TestGormQuery(t *testing.T) {
	db := getPreparedDB(t)
	user := &User{}
	assert.NoError(t, db.Joins("Profile").Where("Profile__age > 1").Find(user).Error)
	fmt.Println(user)
}

type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  string
	FullName  string
	ProfileID uint
	Profile   *UserProfile
}

type UserProfile struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Age       int
}

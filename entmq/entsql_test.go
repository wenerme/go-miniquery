package entmq_test

import (
	"testing"

	"github.com/wenerme/go-miniquery/entmq"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/assert"
)

func TestEntSQL(t *testing.T) {
	for _, test := range []struct {
		E    string
		Q    string
		Args []interface{}
	}{
		{E: `$1 = $2`, Q: `1=1`, Args: []interface{}{1, 1}},
		{E: `"b" < $1 AND ("a" > $2 AND "a" > $3)`, Q: `b < 0 and (a>0 and a > 10)`, Args: []interface{}{0, 0, 10}},
		{E: `"a" < $1 AND "a" > $2`, Q: `a not between 1 and 3`, Args: []interface{}{1, 3}},
		{E: `"a" >= $1 AND "a" <= $2`, Q: `a between 1 and 3`, Args: []interface{}{1, 3}},
		{E: `"a" NOT LIKE $1`, Q: `a not like   "%A%"`, Args: []interface{}{"%A%"}},
		{E: `"a" > $1`, Q: "a > 1", Args: []interface{}{1}},
		{E: `"a" IS NULL`, Q: "a is   null"},
		{E: `"a" IS NULL AND "b" IS NOT NULL`, Q: "a is   null and b is  not  null"},
		{E: `"a" IN ($1, $2, $3)`, Q: "a in [1,2,3]", Args: []interface{}{1, 2, 3}}, // NOTE 目前是处理为多个变量
		{E: `"id" IS NULL`, Q: "ID is   null"},                                      // 目前临时 to_snake_case
		{E: `("activity_type" IN ($1))`, Q: "(activityType in ('PhoneCall'))", Args: []interface{}{"PhoneCall"}},
		{E: `DATE($1)`, Q: `date("2019-01-01 12:12")`, Args: []interface{}{"2019-01-01 12:12"}},
		{E: `DATE("a")`, Q: `date(a)`},
		// 暂不支持
		// {E: `DATE("created_at") between date('2021-05-12T00:00:00+08:00') and date('2021-05-14T00:00:00+08:00')`, Q: `date(created_at) between date('2021-05-12T00:00:00+08:00') and date('2021-05-14T00:00:00+08:00')`},
	} {
		b := &entmq.MiniQLToEntSQLBuilder{QueryString: test.Q, DisableTypeCasting: true}
		b.SetDialect(dialect.Postgres)
		s, args := b.Query()
		assert.NoError(t, b.Err())
		assert.Equal(t, test.E, s)
		assert.EqualValues(t, test.Args, args)
	}
}

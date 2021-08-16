package entmq

import (
	"database/sql/driver"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

func argWithCast(in interface{}) interface{} {
	return typeCastParamFormatter{in}
}

type typeCastParamFormatter struct {
	v interface{}
}

func (f typeCastParamFormatter) FormatParam(placeholder string, info *sql.StmtInfo) string {
	// https://github.com/jackc/pgx/issues/798 pg 需要 cast 来支持 1=1 -> $1=$2
	if info.Dialect == dialect.Postgres {
		var typ string
		switch f.v.(type) {
		case int:
			typ = "int"
		// 字符串不 cast - 处理时间和字符串比较的场景
		// case string:
		//	typ = "text"
		case bool:
			typ = "bool"
		case float64:
			typ = "double precision"
		default:
			return placeholder
		}
		return placeholder + "::" + typ
	}
	return placeholder
}

func (f typeCastParamFormatter) Value() (driver.Value, error) {
	return f.v, nil
}

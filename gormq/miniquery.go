package gormq

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/wenerme/go-miniquery/miniquery"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// MiniQuery Wrap multi miniquery in one scope, will join query by and
type MiniQuery struct {
	Query []string
}

func (q MiniQuery) Scope(db *gorm.DB) *gorm.DB {
	return WireMiniQuery(db, miniquery.Join(q.Query))
}

func GetOrParseSchema(db *gorm.DB) (schema *schema.Schema, err error) {
	stat := db.Statement
	if stat.Schema == nil {
		if stat.Model != nil {
			if err = stat.Parse(stat.Model); err != nil {
				return
			}
		} else {
			return nil, fmt.Errorf("missing model")
		}
	}
	return stat.Schema, nil
}

// ApplyMiniQuery apply single miniquery
func ApplyMiniQuery(query string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return WireMiniQuery(db, query)
	}
}

func WireMiniQuery(db *gorm.DB, query string) *gorm.DB {
	if query == "" {
		return db
	}
	ast, err := miniquery.Parse(query)
	if err != nil {
		_ = db.AddError(fmt.Errorf("invalid query syntax: %v", err))
		return db
	}

	schema, err := GetOrParseSchema(db)
	if err != nil {
		_ = db.AddError(err)
		return db
	}
	var vals []interface{}
	buf := &strings.Builder{}
	joined := map[string][]string{}
	qb := &queryBuilder{
		buf: buf,
		addValue: func(i interface{}) {
			vals = append(vals, i)
		},
		mapName: func(s string) (string, error) {
			n, ok := getDBName(schema, s)
			if !ok {
				return s, fmt.Errorf("field not found: %q", s)
			}
			return n, nil
		},
		join: func(s string, fieldName string) (string, error) {
			r := schema.Relationships.Relations[s]
			if r == nil {
				return "", fmt.Errorf("relation not found: %q", s)
			}

			var name string
			var ok bool
			if name, ok = getDBName(r.FieldSchema, fieldName); !ok {
				return "", fmt.Errorf("relation field not found: %q.%q", s, fieldName)
			}
			if _, ok := joined[s]; !ok {
				db = db.Joins(s)
			}
			joined[s] = append(joined[s], name)
			// gorm Join logic name
			return s + "__" + name, nil
		},
		quote: func(builder *strings.Builder, name string) {
			db.QuoteTo(builder, name)
		},
	}
	err = qb.visit(ast)
	if err != nil {
		_ = db.AddError(err)
	} else {
		db = db.Where(buf.String(), vals...)
	}
	return db
}

type queryBuilder struct {
	buf      *strings.Builder
	addValue func(interface{})
	mapName  func(s string) (string, error)
	quote    func(builder *strings.Builder, name string)
	join     func(s string, f string) (string, error)
}

func (qb *queryBuilder) visitReference(node *miniquery.Node) (err error) {
	if len(node.Names) != 2 {
		return errors.New("only support join one level")
	}
	name, err := qb.join(node.Names[0], node.Names[1])
	if err != nil {
		return err
	}
	quote := qb.quote
	buf := qb.buf
	quote(buf, name)
	return
}

func (qb *queryBuilder) visitIdentifier(node *miniquery.Node) (err error) {
	buf := qb.buf
	mapName := qb.mapName
	quote := qb.quote
	name := node.Name
	/* extension to support custom virtual column */
	// switch name {
	// case "owned":
	// 	buf.WriteString("(owner_id is not null)")
	// 	return
	// case "unowned":
	// 	buf.WriteString("(owner_id is null)")
	// 	return
	// }
	name, err = mapName(name)
	if err != nil {
		return err
	}
	quote(buf, name)
	return
}

func (qb *queryBuilder) visit(node *miniquery.Node) (err error) {
	buf := qb.buf
	addValue := qb.addValue
	visit := qb.visit

	switch node.Type {
	case miniquery.OperationNodeType:
		buf.WriteString(normalizeSQL(node.Operation))
	case miniquery.ValueNodeType:
		// 支持数组值
		buf.WriteRune('?')
		addValue(node.Value())
	case miniquery.IdentifierNodeType:
		err = qb.visitIdentifier(node)
	case miniquery.ReferenceNodeType:
		err = qb.visitReference(node)
	case miniquery.ParenthesesExpressionType:
		buf.WriteRune('(')
		err = visit(node.Expression)
		buf.WriteRune(')')
	case miniquery.NotExpressionType:
		buf.WriteString("not ")
		err = visit(node.Expression)
	case miniquery.FunctionExpressionType:
		err = qb.visitFunction(node)
	case miniquery.BetweenExpressionType:
		err = visit(node.Left)
		if err == nil {
			buf.WriteRune(' ')
			err = visit(node.Op)
		}
		if err == nil {
			buf.WriteRune(' ')
			err = visit(node.Params[0])
		}
		if err == nil {
			buf.WriteString(" and ")
			err = visit(node.Params[1])
		}
	case miniquery.PredicatesExpressionType:
		fallthrough
	case miniquery.LogicExpressionType:
		fallthrough
	case miniquery.CompareExpressionType:
		err = visit(node.Left)
		if err == nil {
			buf.WriteRune(' ')
			err = visit(node.Op)
		}
		if err == nil && node.Right != nil {
			buf.WriteRune(' ')
			err = visit(node.Right)
		}
	default:
		return errors.Errorf("invalid type %q", node.Type)
	}
	return err
}

func (qb *queryBuilder) visitFunction(node *miniquery.Node) (err error) {
	buf := qb.buf
	visit := qb.visit
	switch node.Name {
	case "date":
		buf.WriteString(node.Name)
		buf.WriteString("(")
		for _, v := range node.Params {
			if err = visit(v); err != nil {
				return
			}
		}
		buf.WriteString(")")
	default:
		err = fmt.Errorf("unsupported function: %q", node.Name)
	}
	return
}

func getDBName(st *schema.Schema, name string) (string, bool) {
	if _, ok := st.FieldsByDBName[name]; ok {
		return name, true
	}
	if f, ok := st.FieldsByName[name]; ok {
		return f.DBName, true
	}
	for _, f := range st.Fields {
		if strings.EqualFold(f.Name, name) {
			return f.DBName, true
		}
	}
	return name, false
}

func normalizeSQL(s string) string {
	if v, ok := normalize[s]; ok {
		return v
	}
	return s
}

var normalize = map[string]string{
	"gt":  ">",
	"gte": ">=",
	"eq":  "=",
	"neq": "<>",
	"lt":  "<",
	"lte": "<=",
	"and": "and",
	"or":  "or",
}

package entmq

import (
	"reflect"
	"strings"
	"unsafe"

	"github.com/wenerme/go-miniquery/miniquery"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/huandu/xstrings"
	"github.com/pkg/errors"
)

type MiniQLToEntSQLBuilder struct {
	Node        *sqlgraph.Node
	QueryString string
	ast         *miniquery.Node
	SQLBuilder  *sql.Builder
	Graph       *sqlgraph.Schema
	sql.Builder
	DisableTypeCasting bool
}

// Query impl sql.Querier
func (mb *MiniQLToEntSQLBuilder) Query() (string, []interface{}) {
	if mb.ast == nil {
		ast, err := miniquery.Parse(mb.QueryString)
		if err != nil {
			mb.AddError(err)
			return "", nil
		}
		mb.ast = ast
	}
	err := mb.visit(mb.ast)
	if err != nil {
		mb.AddError(err)
		return "", nil
	}
	return mb.Builder.Query()
}

//nolint:golint,gocyclo
func (mb *MiniQLToEntSQLBuilder) visit(node *miniquery.Node) (err error) {
	visit := mb.visit
	s := mb.SQLBuilder
	if s == nil {
		s = &mb.Builder
	}
	switch node.Type {
	case miniquery.OperationNodeType:
		return errors.Errorf("unexpected op node: %q", node)
	case miniquery.ValueNodeType:
		if node.ValueType == miniquery.ArrayValueType {
			s.WriteString("(")
			s.Args(node.Value().([]interface{})...)
			s.WriteString(")")
		} else {
			// s.Arg(typeCastParamFormatter{node.Value()})
			if mb.DisableTypeCasting {
				s.Arg(node.Value())
			} else {
				s.Arg(argWithCast(node.Value()))
			}
		}
	case miniquery.IdentifierNodeType:
		// fixme 检测字段是否存在
		s.Ident(xstrings.ToSnakeCase(node.Name))
		/*
			switch name {
			case "owned":
				buf.WriteString("(owner_id is not null)")
				return
			case "unowned":
				buf.WriteString("(owner_id is null)")
				return
			}
		*/
	case miniquery.ParenthesesExpressionType:
		s.WriteString("(")
		err = visit(node.Expression)
		s.WriteString(")")
	case miniquery.NotExpressionType:
		s.WriteString("NOT ")
		err = visit(node.Expression)
	case miniquery.FunctionExpressionType:
		// err = qb.visitFunction(node)
		switch node.Name {
		case "date":
			// 截取为日期
			// date(xxx)
			s.WriteString("DATE(")
			err = mb.visit(node.Params[0])
			s.WriteString(")")
		case "has_edge":

			params := node.Params
			if len(params) == 0 || params[0].Name == "" {
				err = errors.New("关联参数错误")
				break
			}

			tab := mb.Node
			builder := sql.Dialect(mb.Dialect()).Select().From(sql.Table(tab.Table).Schema(tab.Schema))
			edge, ok := tab.Edges[params[0].Name]
			if !ok {
				err = errors.Errorf("不存在关联: %q", params[0].Name)
				break
			}
			switch len(params) {
			case 1:
				sqlgraph.HasNeighbors(builder, sqlgraph.NewStep(
					sqlgraph.From(tab.Table, tab.ID.Column),
					sqlgraph.To(edge.To.Table, edge.To.ID.Column),
					sqlgraph.Edge(edge.Spec.Rel, edge.Spec.Inverse, edge.Spec.Table, edge.Spec.Columns...),
				))
			case 2:
				if !params[1].IsExpression() {
					err = errors.Errorf("关联条件错误: %q", params[1].String())
					break
				}
				sqlgraph.HasNeighborsWith(builder, sqlgraph.NewStep(
					sqlgraph.From(tab.Table, tab.ID.Column),
					sqlgraph.To(edge.To.Table, edge.To.ID.Column),
					sqlgraph.Edge(edge.Spec.Rel, edge.Spec.Inverse, edge.Spec.Table, edge.Spec.Columns...),
				), func(selector *sql.Selector) {
					selector.Where(sql.P(func(builder *sql.Builder) {
						mb := &MiniQLToEntSQLBuilder{
							ast:                params[1],
							Node:               edge.To,
							DisableTypeCasting: mb.DisableTypeCasting,
						}
						builder.Join(mb)
						err := mb.Err()
						if err != nil {
							builder.AddError(err)
						}
					}))
				})
			default:
				err = errors.New("关联参数错误")
			}
			if err == nil {
				// hack reflect access
				rv := reflect.ValueOf(builder).Elem()
				rf := rv.FieldByName("where")
				rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
				p := rf.Interface().(*sql.Predicate)
				err = p.Err()
				mb.Join(p)
			}
		default:
			err = errors.Errorf("不支持的函数: %q", node.Name)
		}
	case miniquery.BetweenExpressionType:
		lo := sql.OpGTE
		ro := sql.OpLTE
		if node.Op.Operation == miniquery.OpNotBetween {
			lo = sql.OpLT
			ro = sql.OpGT
		}

		err = visit(node.Left)
		if err == nil {
			s.WriteOp(lo)
			err = visit(node.Params[0])
		}
		if err == nil {
			s.WriteString(" AND ")
		}
		if err == nil {
			err = visit(node.Left)
		}
		if err == nil {
			s.WriteOp(ro)
			err = visit(node.Params[1])
		}
	case miniquery.PredicatesExpressionType:
		op, found := entsqlOpMap[node.Op.Operation]
		if !found {
			return errors.Errorf("unexpected predicate op %q", node.Op.Operation)
		}
		err = visit(node.Left)
		s.WriteOp(op)
	case miniquery.LogicExpressionType:
		fallthrough
	case miniquery.CompareExpressionType:
		err = visit(node.Left)
		op, found := entsqlOpMap[node.Op.Operation]
		if !found {
			switch node.Op.Operation {
			case miniquery.OpNotLike:
				fallthrough
			case miniquery.OpAnd:
				fallthrough
			case miniquery.OpOr:
				s.Pad().WriteString(strings.ToUpper(node.Op.Operation)).Pad()
				found = true
			}
		} else {
			s.WriteOp(op)
		}
		if !found {
			return errors.Errorf("unexpected op %q", node.Op.Operation)
		}
		if err == nil {
			err = visit(node.Right)
		}
	default:
		return errors.Errorf("invalid type %q", node.Type)
	}
	return err
}

var entsqlOpMap = map[miniquery.OpType]sql.Op{
	miniquery.OpEQ:        sql.OpEQ,
	miniquery.OpNEQ:       sql.OpNEQ,
	miniquery.OpGT:        sql.OpGT,
	miniquery.OpGTE:       sql.OpGTE,
	miniquery.OpLT:        sql.OpLT,
	miniquery.OpLTE:       sql.OpLTE,
	miniquery.OpIn:        sql.OpIn,
	miniquery.OpNotIn:     sql.OpNotIn,
	miniquery.OpLike:      sql.OpLike,
	miniquery.OpIsNull:    sql.OpIsNull,
	miniquery.OpIsNotNull: sql.OpNotNull,
}

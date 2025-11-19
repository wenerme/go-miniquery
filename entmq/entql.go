package entmq

import (
	"log/slog"

	"entgo.io/ent/entql"
	"github.com/pkg/errors"
	"github.com/wenerme/go-miniquery/miniquery"
)

func ApplyEntQL(q interface{ Where(p entql.P) }, queries []string) error {
	for _, v := range queries {
		mb := MiniQLToEntQLBuilder{Query: v}
		p, err := mb.Build()
		if err != nil {
			return errors.Wrapf(err, "failed to apply query: %q", v)
		}
		q.Where(p)
	}
	return nil
}

func BuildEntQL(v string) (entql.P, error) {
	mb := MiniQLToEntQLBuilder{Query: v}
	p, err := mb.Build()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build query: %q", v)
	}
	return p, nil
}

type MiniQLToEntQLBuilder struct {
	Query string
	stack []entql.Expr
}

func (mb *MiniQLToEntQLBuilder) pop() entql.Expr {
	v := mb.stack[len(mb.stack)-1]
	mb.stack = mb.stack[:len(mb.stack)-1]
	return v
}

func (mb *MiniQLToEntQLBuilder) push(e entql.Expr) {
	mb.stack = append(mb.stack, e)
}

func (mb *MiniQLToEntQLBuilder) Build() (p entql.P, err error) {
	if mb.Query == "" {
		return
	}
	node, err := miniquery.Parse(mb.Query)
	if err != nil {
		return nil, err
	}
	err = mb.visit(node)
	if err == nil && len(mb.stack) > 0 {
		p = mb.pop().(entql.P)
		if len(mb.stack) > 1 {
			slog.With("query", mb.Query).Error("unexpected MiniQLBuilder stack size", "size", len(mb.stack))
		}
	}
	return
}

func (mb *MiniQLToEntQLBuilder) visit(node *miniquery.Node) (err error) {
	visit := mb.visit

	switch node.Type {
	case miniquery.OperationNodeType:
		return errors.Errorf("unexpected op node: %q", node)
	case miniquery.ValueNodeType:
		mb.push(&entql.Value{V: node.Value()})
	case miniquery.IdentifierNodeType:
		mb.push(entql.F(node.Name))
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
		err = visit(node.Expression)
	case miniquery.NotExpressionType:
		err = visit(node.Expression)
		if err != nil {
			mb.push(entql.Not(mb.pop().(entql.P)))
		}
	case miniquery.FunctionExpressionType:
		// err = qb.visitFunction(node)

	case miniquery.BetweenExpressionType:
		err = visit(node.Left)
		if err == nil {
			err = visit(node.Params[0])
		}
		if err == nil {
			err = visit(node.Params[1])
		}
		if err == nil {
			hig := mb.pop()
			low := mb.pop()
			left := mb.pop()
			lo := entql.OpGTE
			ro := entql.OpLTE
			if node.Op.Operation == miniquery.OpNotBetween {
				lo = entql.OpLT
				ro = entql.OpGT
			}
			mb.push(entql.And(
				&entql.BinaryExpr{
					Op: lo,
					X:  left,
					Y:  low,
				},
				&entql.BinaryExpr{
					Op: ro,
					X:  left,
					Y:  hig,
				},
			))

		}
	case miniquery.PredicatesExpressionType:
		panic("TODO")
	case miniquery.LogicExpressionType:
		fallthrough
	case miniquery.CompareExpressionType:
		err = visit(node.Left)
		if err == nil {
			err = visit(node.Right)
		}
		if err == nil {
			right := mb.pop()
			left := mb.pop()
			op, found := entqlOpMap[node.Op.Operation]
			if !found {
				return errors.Errorf("unexpected op %q", node.Op.Operation)
			}
			mb.push(&entql.BinaryExpr{
				Op: op,
				X:  left,
				Y:  right,
			})
		}
	default:
		return errors.Errorf("invalid type %q", node.Type)
	}
	return err
}

var entqlOpMap = map[miniquery.OpType]entql.Op{
	miniquery.OpAnd:   entql.OpAnd,
	miniquery.OpOr:    entql.OpOr,
	miniquery.OpNot:   entql.OpNot,
	miniquery.OpEQ:    entql.OpEQ,
	miniquery.OpNEQ:   entql.OpNEQ,
	miniquery.OpGT:    entql.OpGT,
	miniquery.OpGTE:   entql.OpGTE,
	miniquery.OpLT:    entql.OpLT,
	miniquery.OpLTE:   entql.OpLTE,
	miniquery.OpIn:    entql.OpIn,
	miniquery.OpNotIn: entql.OpNotIn,
}

package miniquery

import (
	"context"

	"github.com/pkg/errors"
)

type (
	VisitFunc = func(ctx context.Context, node *Node) error
	Visitor   struct {
		BinaryExpression      VisitFunc
		LogicExpression       VisitFunc
		BetweenExpression     VisitFunc
		FunctionExpression    VisitFunc
		PredicatesExpression  VisitFunc
		TerminateExpression   VisitFunc
		ParenthesesExpression VisitFunc
		NotExpression         VisitFunc
		IdentifierNode        VisitFunc
		OperationNode         VisitFunc
		ValueNode             VisitFunc
	}
)

func (v Visitor) Init() {
	noop := func(ctx context.Context, node *Node) (err error) {
		return
	}
	if v.OperationNode == nil {
		v.OperationNode = noop
	}
	if v.IdentifierNode == nil {
		v.IdentifierNode = noop
	}
	if v.ParenthesesExpression == nil {
		v.ParenthesesExpression = func(ctx context.Context, node *Node) error {
			return v.Visit(ctx, node.Expression)
		}
	}
	if v.NotExpression == nil {
		v.NotExpression = func(ctx context.Context, node *Node) error {
			return v.Visit(ctx, node.Expression)
		}
	}
	if v.FunctionExpression == nil {
		v.FunctionExpression = func(ctx context.Context, node *Node) (err error) {
			for _, n := range node.Params {
				if err = v.Visit(ctx, n); err != nil {
					return
				}
			}
			return
		}
	}
	binary := func(ctx context.Context, node *Node) (err error) {
		err = v.Visit(ctx, node.Left)
		if err == nil {
			err = v.Visit(ctx, node.Op)
		}
		if err == nil {
			err = v.Visit(ctx, node.Right)
		}
		return
	}
	if v.BinaryExpression == nil {
		v.BinaryExpression = binary
	}
	if v.LogicExpression == nil {
		v.LogicExpression = binary
	}
	if v.PredicatesExpression == nil {
		v.PredicatesExpression = func(ctx context.Context, node *Node) (err error) {
			err = v.Visit(ctx, node.Left)
			if err == nil {
				err = v.Visit(ctx, node.Op)
			}
			return
		}
	}
	if v.BetweenExpression == nil {
		v.BetweenExpression = func(ctx context.Context, node *Node) (err error) {
			err = v.Visit(ctx, node.Left)
			if err == nil {
				err = v.Visit(ctx, node.Op)
			}
			if err == nil {
				for _, n := range node.Params {
					if err = v.Visit(ctx, n); err != nil {
						return
					}
				}
			}
			return
		}
	}
	if v.ValueNode == nil {
		v.ValueNode = func(ctx context.Context, node *Node) (err error) {
			if node.ValueType == ArrayValueType {
				for _, n := range node.Array {
					if err = v.Visit(ctx, n); err != nil {
						return
					}
				}
			}
			return
		}
	}
}

func (v Visitor) Visit(ctx context.Context, n *Node) error {
	switch n.Type {
	case ValueNodeType:
		return v.ValueNode(ctx, n)
	case OperationNodeType:
		return v.OperationNode(ctx, n)
	case IdentifierNodeType:
		return v.IdentifierNode(ctx, n)
	case ParenthesesExpressionType:
		return v.ParenthesesExpression(ctx, n)
	case CompareExpressionType:
		return v.BinaryExpression(ctx, n)
	case LogicExpressionType:
		return v.LogicExpression(ctx, n)
	case PredicatesExpressionType:
		return v.PredicatesExpression(ctx, n)
	case NotExpressionType:
		return v.NotExpression(ctx, n)
	case BetweenExpressionType:
		return v.BetweenExpression(ctx, n)
	case FunctionExpressionType:
		return v.BetweenExpression(ctx, n)
	}
	return errors.Errorf("invalid type %q", n.Type)
}

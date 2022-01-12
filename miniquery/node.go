package miniquery

import (
	"fmt"
	"strings"
)

type (
	NodeType  string
	ValueType string
)

const (
	ValueNodeType             NodeType = "value"       // Node.Value
	OperationNodeType         NodeType = "operation"   // Node.Operation
	IdentifierNodeType        NodeType = "identifier"  // Node.Name
	ReferenceNodeType         NodeType = "reference"   // Node.Names
	PredicatesExpressionType  NodeType = "predicates"  // field is null - Node.Left, Node.Op
	CompareExpressionType     NodeType = "compare"     // Node.Left, Node.Op, Node.Right
	LogicExpressionType       NodeType = "logic"       // Node.Left, Node.Op, Node.Right
	ParenthesesExpressionType NodeType = "parentheses" // Node.Expression
	NotExpressionType         NodeType = "not"         // Node.Expression
	BetweenExpressionType     NodeType = "between"     // Node.Left, Node.Op, Node.Params
	FunctionExpressionType    NodeType = "function"    // Node.Name, Node.Params
)

const (
	IntValueType     ValueType = "int"
	FloatValueType   ValueType = "float"
	BooleanValueType ValueType = "bool"
	StringValueType  ValueType = "string"
	NullValueType    ValueType = "null"
	ArrayValueType   ValueType = "array"
)

type Node struct {
	Type      NodeType
	ValueType ValueType
	Int       int
	Bool      bool
	Str       string
	Float     float64
	Array     []*Node

	Operation OpType
	Name      string
	Names     []string // Reference

	Expression *Node

	Left   *Node
	Op     *Node // Operation
	Right  *Node
	Params []*Node
}

func (n Node) Value() interface{} {
	switch n.ValueType {
	case IntValueType:
		return n.Int
	case StringValueType:
		return n.Str
	case FloatValueType:
		return n.Float
	case BooleanValueType:
		return n.Bool
	case NullValueType:
		return nil
	case ArrayValueType:
		var s []interface{}
		for _, v := range n.Array {
			s = append(s, v.Value())
		}
		return s
	}
	return "<unknown value>"
}

func (n Node) IsValue() bool {
	return n.ValueType != ""
}

func (n Node) IsExpression() bool {
	switch n.Type {
	case IdentifierNodeType, ValueNodeType, OperationNodeType, ReferenceNodeType:
		return false
	}
	return true
}

func (n Node) String() string {
	buf := &strings.Builder{}
	buf.WriteString(string(n.Type))
	buf.WriteString("(")

	switch n.Type {
	case BetweenExpressionType:
		buf.WriteString(n.Params[0].String())
		buf.WriteString(",")
		buf.WriteString(n.Params[1].String())
	case PredicatesExpressionType:
		buf.WriteString(n.Operation)
	case LogicExpressionType:
		buf.WriteString(n.Op.String())
		buf.WriteString(",")
		buf.WriteString(n.Left.String())
		buf.WriteString(",")
		buf.WriteString(n.Right.String())
	case CompareExpressionType:
		buf.WriteString(n.Left.String())
		buf.WriteString(",")
		buf.WriteString(n.Op.String())
		buf.WriteString(",")
		buf.WriteString(n.Right.String())
	case ParenthesesExpressionType:
		fallthrough
	case NotExpressionType:
		buf.WriteString(n.Expression.String())
	case ValueNodeType:
		buf.WriteString(fmt.Sprint(n.Value()))
	case OperationNodeType:
		buf.WriteString(n.Operation)
	case IdentifierNodeType:
		buf.WriteString(n.Name)
	case ReferenceNodeType:
		buf.WriteString(strings.Join(n.Names, "."))
	case FunctionExpressionType:
		buf.WriteString(n.Name)
		for _, v := range n.Params {
			buf.WriteString(",")
			buf.WriteString(v.String())
		}
	default:
		buf.WriteString("unknown")
	}
	buf.WriteString(")")
	return buf.String()
}

func Build(n *Node) string {
	buf := &strings.Builder{}
	var visit func(node *Node)
	visit = func(node *Node) {
		switch node.Type {
		case NotExpressionType:
			buf.WriteString("not ")
			visit(node.Expression)
		case BetweenExpressionType:
			visit(node.Left)
			buf.WriteString(" ")
			visit(node.Op)
			buf.WriteString(" ")
			visit(node.Params[0])
			buf.WriteString(" and ")
			visit(node.Params[1])
		case PredicatesExpressionType:
			fallthrough
		case LogicExpressionType:
			fallthrough
		case CompareExpressionType:
			visit(node.Left)
			buf.WriteRune(' ')
			visit(node.Op)
			if node.Right != nil {
				buf.WriteRune(' ')
				visit(node.Right)
			}
		case ParenthesesExpressionType:
			buf.WriteRune('(')
			visit(node.Expression)
			buf.WriteRune(')')
		case OperationNodeType:
			buf.WriteString(printPretty(node.Operation))
		case IdentifierNodeType:
			buf.WriteString(node.Name)
		case ReferenceNodeType:
			buf.WriteString(strings.Join(node.Names, "."))
		case FunctionExpressionType:
			buf.WriteString(n.Name)
			buf.WriteRune('(')
			n := len(node.Params)
			for i, v := range node.Params {
				visit(v)
				if i != n-1 {
					buf.WriteRune(',')
				}
			}
			buf.WriteRune(')')
		case ValueNodeType:
			buildValue(buf, node, visit)
		default:
			buf.WriteString("<unknown type>")
		}
	}
	visit(n)
	return buf.String()
}

func buildValue(buf *strings.Builder, node *Node, visit func(node *Node)) {
	switch node.ValueType {
	case ArrayValueType:
		buf.WriteRune('[')
		n := len(node.Array)
		for i, v := range node.Array {
			visit(v)
			if i != n-1 {
				buf.WriteRune(',')
			}
		}
		buf.WriteRune(']')
	case StringValueType:
		buf.WriteString(fmt.Sprintf("%q", node.Str))
	default:
		value := node.Value()
		if value == nil {
			buf.WriteString("null")
		} else {
			buf.WriteString(fmt.Sprint(value))
		}
	}
}

type OpType = string

const (
	OpGTE        OpType = "gte"
	OpGT         OpType = "gt"
	OpEQ         OpType = "eq"
	OpNEQ        OpType = "neq"
	OpLT         OpType = "lt"
	OpLTE        OpType = "lte"
	OpAnd        OpType = "and"
	OpOr         OpType = "or"
	OpBetween    OpType = "between"
	OpLike       OpType = "like"
	OpNotLike    OpType = "not like"
	OpNotBetween OpType = "not between"
	OpIsNull     OpType = "is null"
	OpIsNotNull  OpType = "is not null"
	OpNot        OpType = "not"
	OpIn         OpType = "in"
	OpNotIn      OpType = "not in"
	// IS UNKNOWN, TRUE, FALSE, DISTINCT FROM
	// BETWEEN SYMMETRIC
)

var pretties = map[string]string{
	"gte":       ">=",
	"gt":        ">",
	"eq":        "==",
	"neq":       "!=",
	"lt":        "<",
	"lte":       "<=",
	"and":       "&&",
	"or":        "||",
	"isnull":    "is null",
	"isnotnull": "is not null",
	"notin":     "not in",
}

func printPretty(s string) string {
	if v, ok := pretties[s]; ok {
		return v
	}
	return s
}

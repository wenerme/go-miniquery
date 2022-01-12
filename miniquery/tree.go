//go:generate peg -switch miniquery.peg
package miniquery

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Tree build time syntax tree
//
// reference
// https://www.postgresql.org/docs/current/functions-comparison.html
type Tree struct {
	Stack  []*Node
	Errors []error
	marks  []int
}

var (
	regSpace  = regexp.MustCompile(`\s+`)
	normalize = map[string]string{
		">=": "gte",
		">":  "gt",
		"==": "eq",
		"=":  "eq",
		"!=": "neq",
		"<>": "neq",
		"<":  "lt",
		"<=": "lte",
		":":  "eq",
		//
		"&&": "and",
		"||": "or",
		",":  "and",

		// 非标准 sql 语法
		"isnull":  "is null",
		"notnull": "is not null",
	}
)

func NormalizeOperation(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = regSpace.ReplaceAllString(s, " ")
	if v, ok := normalize[s]; ok {
		return v
	}
	return s
}

func (t *Tree) Push(node *Node) {
	t.Stack = append(t.Stack, node)
}

func (t *Tree) Pop() *Node {
	last := len(t.Stack) - 1
	if last < 0 {
		return nil
	}
	v := t.Stack[last]
	t.Stack = t.Stack[:last]
	return v
}

func (t *Tree) PopIdentifierReference() {
	nodes := t.popMarked()
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	t.Push(&Node{
		Type:  ReferenceNodeType,
		Names: names,
	})
}

func (t *Tree) PopFunction() {
	a := t.Pop()
	n := t.Pop()
	if !(n.Type == IdentifierNodeType && a.Type == ValueNodeType && a.ValueType == ArrayValueType) {
		t.AddError(fmt.Errorf("invalid function exp %q %q", n.Type, a.Type))
		return
	}
	t.Push(&Node{
		Type:   FunctionExpressionType,
		Name:   n.Name,
		Params: a.Array,
	})
}

func (t *Tree) PopBetween() {
	up := t.Pop()
	down := t.Pop()
	op := t.Pop()
	left := t.Pop()
	t.Push(&Node{
		Type: BetweenExpressionType,
		Left: left,
		Op:   op,
		Params: []*Node{
			down,
			up,
		},
	})
}

func (t *Tree) PopPredicate() {
	op := t.Pop()
	left := t.Pop()

	t.Push(&Node{
		Type: PredicatesExpressionType,
		Left: left,
		Op:   op,
	})
}

func (t *Tree) PopLogic() {
	right := t.Pop()
	op := t.Pop()
	left := t.Pop()
	t.Push(&Node{
		Type:  LogicExpressionType,
		Left:  left,
		Op:    op,
		Right: right,
	})
}

func (t *Tree) PopCompare() {
	right := t.Pop()
	op := t.Pop()
	left := t.Pop()

	t.Push(&Node{
		Type:  CompareExpressionType,
		Left:  left,
		Op:    op,
		Right: right,
	})
}

func (t *Tree) PopParentheses() {
	t.Push(&Node{
		Type:       ParenthesesExpressionType,
		Expression: t.Pop(),
	})
}

func (t *Tree) PopMark() int {
	v := t.marks[len(t.marks)-1]
	t.marks = t.marks[:len(t.marks)-1]
	return v
}

func (t *Tree) AddMark() {
	t.marks = append(t.marks, len(t.Stack))
}

func (t *Tree) popMarked() []*Node {
	mark := t.PopMark()
	elements := make([]*Node, len(t.Stack)-mark)
	copy(elements, t.Stack[mark:len(t.Stack)])
	t.Stack = t.Stack[0:mark]
	return elements
}

func (t *Tree) PopArray() {
	t.Push(&Node{
		Type:      ValueNodeType,
		ValueType: ArrayValueType,
		Array:     t.popMarked(),
	})
}

func (t *Tree) AddOperation(s string) {
	t.Push(&Node{
		Type:      OperationNodeType,
		Operation: NormalizeOperation(s),
	})
}

func (t *Tree) AddCompare(s string) {
	t.Push(&Node{
		Type:      OperationNodeType,
		Operation: NormalizeOperation(s),
	})
}

func (t *Tree) AddLogic(s string) {
	t.Push(&Node{
		Type:      OperationNodeType,
		Operation: NormalizeOperation(s),
	})
}

func (t *Tree) PopNot() {
	t.Push(&Node{
		Type:       NotExpressionType,
		Expression: t.Pop(),
	})
}

func (t *Tree) AddMatch(s string) {
	t.Push(&Node{
		Type:      OperationNodeType,
		Operation: NormalizeOperation(s),
	})
}

func (t *Tree) AddName(s string) {
	t.Push(&Node{
		Type: IdentifierNodeType,
		Name: s,
	})
}

func (t *Tree) AddError(err error) {
	if err != nil {
		t.Errors = append(t.Errors, err)
	}
}

func (t *Tree) AddString(s string) {
	t.Stack = append(t.Stack, &Node{
		Type:      ValueNodeType,
		ValueType: StringValueType,
		Str:       s,
	})
}

func (t *Tree) AddFloat(s string) {
	v, err := strconv.ParseFloat(s, 64)
	t.AddError(err)
	t.Stack = append(t.Stack, &Node{
		Type:      ValueNodeType,
		ValueType: FloatValueType,
		Float:     v,
	})
}

func (t *Tree) AddNull() {
	t.Stack = append(t.Stack, &Node{
		Type:      ValueNodeType,
		ValueType: NullValueType,
	})
}

func (t *Tree) AddBoolean(s string) {
	i, err := strconv.ParseBool(s)
	t.AddError(err)
	t.Stack = append(t.Stack, &Node{
		Type:      ValueNodeType,
		ValueType: BooleanValueType,
		Bool:      i,
	})
}

func (t *Tree) AddInteger(s string) {
	i, err := strconv.Atoi(s)
	t.AddError(err)
	t.Stack = append(t.Stack, &Node{
		Type:      ValueNodeType,
		ValueType: IntValueType,
		Int:       i,
	})
}

package miniquery

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleGrammar
	ruleExpression
	ruleLogicExpression
	ruleNotExpression
	ruleCompareExpression
	ruleCompareInExpression
	rulePredicateExpression
	ruleBetweenExpression
	rulePrimaryExpression
	ruleArgumentList
	ruleArgument
	ruleReference
	ruleJsonReference
	ruleIdentifier
	ruleCompare
	ruleLogic
	ruleMatch
	ruleValue
	ruleArray
	ruleLiteral
	ruleInteger
	ruleBoolean
	ruleNull
	ruleString
	ruleSpaceComment
	rule_
	rule__
	ruleComment
	ruleSpace
	ruleEndOfLine
	ruleEndOfFile
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25
	ruleAction26
	ruleAction27

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"Grammar",
	"Expression",
	"LogicExpression",
	"NotExpression",
	"CompareExpression",
	"CompareInExpression",
	"PredicateExpression",
	"BetweenExpression",
	"PrimaryExpression",
	"ArgumentList",
	"Argument",
	"Reference",
	"JsonReference",
	"Identifier",
	"Compare",
	"Logic",
	"Match",
	"Value",
	"Array",
	"Literal",
	"Integer",
	"Boolean",
	"Null",
	"String",
	"SpaceComment",
	"_",
	"__",
	"Comment",
	"Space",
	"EndOfLine",
	"EndOfFile",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",
	"Action26",
	"Action27",

	"Pre_",
	"_In_",
	"_Suf",
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens32) Expand(index int) {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
}

type MiniQueryPeg struct {
	*Tree

	Buffer string
	buffer []rune
	rules  [61]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokens32
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *MiniQueryPeg
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *MiniQueryPeg) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *MiniQueryPeg) Highlighter() {
	p.PrintSyntax()
}

func (p *MiniQueryPeg) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.PopLogic()
		case ruleAction1:
			p.PopNot()
		case ruleAction2:
			p.PopCompare()
		case ruleAction3:
			p.AddCompare(text)
		case ruleAction4:
			p.PopCompare()
		case ruleAction5:
			p.PopPredicate()
		case ruleAction6:
			p.AddOperation(text)
		case ruleAction7:
			p.PopBetween()
		case ruleAction8:
			p.PopParentheses()
		case ruleAction9:
			p.PopFunction()
		case ruleAction10:
			p.AddMark()
		case ruleAction11:
			p.PopArray()
		case ruleAction12:
			p.AddName(text)
		case ruleAction13:
			p.AddCompare(text)
		case ruleAction14:
			p.AddCompare(text)
		case ruleAction15:
			p.AddCompare(text)
		case ruleAction16:
			p.AddLogic(text)
		case ruleAction17:
			p.AddLogic(text)
		case ruleAction18:
			p.AddMatch(text)
		case ruleAction19:
			p.AddMark()
		case ruleAction20:
			p.PopArray()
		case ruleAction21:
			p.AddMark()
		case ruleAction22:
			p.PopArray()
		case ruleAction23:
			p.AddInteger(text)
		case ruleAction24:
			p.AddBoolean(text)
		case ruleAction25:
			p.AddNull()
		case ruleAction26:
			p.AddString(text)
		case ruleAction27:
			p.AddString(text)

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *MiniQueryPeg) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		tree.Expand(tokenIndex)
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Grammar <- <(_ Expression _ EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[rule_]() {
					goto l0
				}
				if !_rules[ruleExpression]() {
					goto l0
				}
				if !_rules[rule_]() {
					goto l0
				}
				if !_rules[ruleEndOfFile]() {
					goto l0
				}
				depth--
				add(ruleGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Expression <- <LogicExpression> */
		func() bool {
			position2, tokenIndex2, depth2 := position, tokenIndex, depth
			{
				position3 := position
				depth++
				if !_rules[ruleLogicExpression]() {
					goto l2
				}
				depth--
				add(ruleExpression, position3)
			}
			return true
		l2:
			position, tokenIndex, depth = position2, tokenIndex2, depth2
			return false
		},
		/* 2 LogicExpression <- <(NotExpression (Logic NotExpression Action0)*)> */
		func() bool {
			position4, tokenIndex4, depth4 := position, tokenIndex, depth
			{
				position5 := position
				depth++
				if !_rules[ruleNotExpression]() {
					goto l4
				}
			l6:
				{
					position7, tokenIndex7, depth7 := position, tokenIndex, depth
					if !_rules[ruleLogic]() {
						goto l7
					}
					if !_rules[ruleNotExpression]() {
						goto l7
					}
					if !_rules[ruleAction0]() {
						goto l7
					}
					goto l6
				l7:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
				}
				depth--
				add(ruleLogicExpression, position5)
			}
			return true
		l4:
			position, tokenIndex, depth = position4, tokenIndex4, depth4
			return false
		},
		/* 3 NotExpression <- <(CompareExpression / (_ (('n' / 'N') ('o' / 'O') ('t' / 'T')) __ CompareExpression Action1))> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				{
					position10, tokenIndex10, depth10 := position, tokenIndex, depth
					if !_rules[ruleCompareExpression]() {
						goto l11
					}
					goto l10
				l11:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if !_rules[rule_]() {
						goto l8
					}
					{
						position12, tokenIndex12, depth12 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l13
						}
						position++
						goto l12
					l13:
						position, tokenIndex, depth = position12, tokenIndex12, depth12
						if buffer[position] != rune('N') {
							goto l8
						}
						position++
					}
				l12:
					{
						position14, tokenIndex14, depth14 := position, tokenIndex, depth
						if buffer[position] != rune('o') {
							goto l15
						}
						position++
						goto l14
					l15:
						position, tokenIndex, depth = position14, tokenIndex14, depth14
						if buffer[position] != rune('O') {
							goto l8
						}
						position++
					}
				l14:
					{
						position16, tokenIndex16, depth16 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l17
						}
						position++
						goto l16
					l17:
						position, tokenIndex, depth = position16, tokenIndex16, depth16
						if buffer[position] != rune('T') {
							goto l8
						}
						position++
					}
				l16:
					if !_rules[rule__]() {
						goto l8
					}
					if !_rules[ruleCompareExpression]() {
						goto l8
					}
					if !_rules[ruleAction1]() {
						goto l8
					}
				}
			l10:
				depth--
				add(ruleNotExpression, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 4 CompareExpression <- <(CompareInExpression (Compare CompareInExpression Action2)*)> */
		func() bool {
			position18, tokenIndex18, depth18 := position, tokenIndex, depth
			{
				position19 := position
				depth++
				if !_rules[ruleCompareInExpression]() {
					goto l18
				}
			l20:
				{
					position21, tokenIndex21, depth21 := position, tokenIndex, depth
					if !_rules[ruleCompare]() {
						goto l21
					}
					if !_rules[ruleCompareInExpression]() {
						goto l21
					}
					if !_rules[ruleAction2]() {
						goto l21
					}
					goto l20
				l21:
					position, tokenIndex, depth = position21, tokenIndex21, depth21
				}
				depth--
				add(ruleCompareExpression, position19)
			}
			return true
		l18:
			position, tokenIndex, depth = position18, tokenIndex18, depth18
			return false
		},
		/* 5 CompareInExpression <- <(PredicateExpression (_ <((('i' / 'I') ('n' / 'N')) / (('n' / 'N') ('o' / 'O') ('t' / 'T') __ (('i' / 'I') ('n' / 'N'))))> _ Action3 Array Action4)?)> */
		func() bool {
			position22, tokenIndex22, depth22 := position, tokenIndex, depth
			{
				position23 := position
				depth++
				if !_rules[rulePredicateExpression]() {
					goto l22
				}
				{
					position24, tokenIndex24, depth24 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l24
					}
					{
						position26 := position
						depth++
						{
							position27, tokenIndex27, depth27 := position, tokenIndex, depth
							{
								position29, tokenIndex29, depth29 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l30
								}
								position++
								goto l29
							l30:
								position, tokenIndex, depth = position29, tokenIndex29, depth29
								if buffer[position] != rune('I') {
									goto l28
								}
								position++
							}
						l29:
							{
								position31, tokenIndex31, depth31 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l32
								}
								position++
								goto l31
							l32:
								position, tokenIndex, depth = position31, tokenIndex31, depth31
								if buffer[position] != rune('N') {
									goto l28
								}
								position++
							}
						l31:
							goto l27
						l28:
							position, tokenIndex, depth = position27, tokenIndex27, depth27
							{
								position33, tokenIndex33, depth33 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l34
								}
								position++
								goto l33
							l34:
								position, tokenIndex, depth = position33, tokenIndex33, depth33
								if buffer[position] != rune('N') {
									goto l24
								}
								position++
							}
						l33:
							{
								position35, tokenIndex35, depth35 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l36
								}
								position++
								goto l35
							l36:
								position, tokenIndex, depth = position35, tokenIndex35, depth35
								if buffer[position] != rune('O') {
									goto l24
								}
								position++
							}
						l35:
							{
								position37, tokenIndex37, depth37 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l38
								}
								position++
								goto l37
							l38:
								position, tokenIndex, depth = position37, tokenIndex37, depth37
								if buffer[position] != rune('T') {
									goto l24
								}
								position++
							}
						l37:
							if !_rules[rule__]() {
								goto l24
							}
							{
								position39, tokenIndex39, depth39 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l40
								}
								position++
								goto l39
							l40:
								position, tokenIndex, depth = position39, tokenIndex39, depth39
								if buffer[position] != rune('I') {
									goto l24
								}
								position++
							}
						l39:
							{
								position41, tokenIndex41, depth41 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l42
								}
								position++
								goto l41
							l42:
								position, tokenIndex, depth = position41, tokenIndex41, depth41
								if buffer[position] != rune('N') {
									goto l24
								}
								position++
							}
						l41:
						}
					l27:
						depth--
						add(rulePegText, position26)
					}
					if !_rules[rule_]() {
						goto l24
					}
					if !_rules[ruleAction3]() {
						goto l24
					}
					if !_rules[ruleArray]() {
						goto l24
					}
					if !_rules[ruleAction4]() {
						goto l24
					}
					goto l25
				l24:
					position, tokenIndex, depth = position24, tokenIndex24, depth24
				}
			l25:
				depth--
				add(ruleCompareInExpression, position23)
			}
			return true
		l22:
			position, tokenIndex, depth = position22, tokenIndex22, depth22
			return false
		},
		/* 6 PredicateExpression <- <(BetweenExpression (Match Action5)?)> */
		func() bool {
			position43, tokenIndex43, depth43 := position, tokenIndex, depth
			{
				position44 := position
				depth++
				if !_rules[ruleBetweenExpression]() {
					goto l43
				}
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if !_rules[ruleMatch]() {
						goto l45
					}
					if !_rules[ruleAction5]() {
						goto l45
					}
					goto l46
				l45:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
				}
			l46:
				depth--
				add(rulePredicateExpression, position44)
			}
			return true
		l43:
			position, tokenIndex, depth = position43, tokenIndex43, depth43
			return false
		},
		/* 7 BetweenExpression <- <(PrimaryExpression (_ <((('n' / 'N') ('o' / 'O') ('t' / 'T') __)? ('b' 'e' 't' 'w' 'e' 'e' 'n'))> Action6 _ ((Value _ (('a' / 'A') ('n' / 'N') ('d' / 'D')) _ Value) / ('[' _ Value _ ',' _ Value _ ']')) Action7)?)> */
		func() bool {
			position47, tokenIndex47, depth47 := position, tokenIndex, depth
			{
				position48 := position
				depth++
				if !_rules[rulePrimaryExpression]() {
					goto l47
				}
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l49
					}
					{
						position51 := position
						depth++
						{
							position52, tokenIndex52, depth52 := position, tokenIndex, depth
							{
								position54, tokenIndex54, depth54 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l55
								}
								position++
								goto l54
							l55:
								position, tokenIndex, depth = position54, tokenIndex54, depth54
								if buffer[position] != rune('N') {
									goto l52
								}
								position++
							}
						l54:
							{
								position56, tokenIndex56, depth56 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l57
								}
								position++
								goto l56
							l57:
								position, tokenIndex, depth = position56, tokenIndex56, depth56
								if buffer[position] != rune('O') {
									goto l52
								}
								position++
							}
						l56:
							{
								position58, tokenIndex58, depth58 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l59
								}
								position++
								goto l58
							l59:
								position, tokenIndex, depth = position58, tokenIndex58, depth58
								if buffer[position] != rune('T') {
									goto l52
								}
								position++
							}
						l58:
							if !_rules[rule__]() {
								goto l52
							}
							goto l53
						l52:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
						}
					l53:
						if buffer[position] != rune('b') {
							goto l49
						}
						position++
						if buffer[position] != rune('e') {
							goto l49
						}
						position++
						if buffer[position] != rune('t') {
							goto l49
						}
						position++
						if buffer[position] != rune('w') {
							goto l49
						}
						position++
						if buffer[position] != rune('e') {
							goto l49
						}
						position++
						if buffer[position] != rune('e') {
							goto l49
						}
						position++
						if buffer[position] != rune('n') {
							goto l49
						}
						position++
						depth--
						add(rulePegText, position51)
					}
					if !_rules[ruleAction6]() {
						goto l49
					}
					if !_rules[rule_]() {
						goto l49
					}
					{
						position60, tokenIndex60, depth60 := position, tokenIndex, depth
						if !_rules[ruleValue]() {
							goto l61
						}
						if !_rules[rule_]() {
							goto l61
						}
						{
							position62, tokenIndex62, depth62 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l63
							}
							position++
							goto l62
						l63:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if buffer[position] != rune('A') {
								goto l61
							}
							position++
						}
					l62:
						{
							position64, tokenIndex64, depth64 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l65
							}
							position++
							goto l64
						l65:
							position, tokenIndex, depth = position64, tokenIndex64, depth64
							if buffer[position] != rune('N') {
								goto l61
							}
							position++
						}
					l64:
						{
							position66, tokenIndex66, depth66 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l67
							}
							position++
							goto l66
						l67:
							position, tokenIndex, depth = position66, tokenIndex66, depth66
							if buffer[position] != rune('D') {
								goto l61
							}
							position++
						}
					l66:
						if !_rules[rule_]() {
							goto l61
						}
						if !_rules[ruleValue]() {
							goto l61
						}
						goto l60
					l61:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
						if buffer[position] != rune('[') {
							goto l49
						}
						position++
						if !_rules[rule_]() {
							goto l49
						}
						if !_rules[ruleValue]() {
							goto l49
						}
						if !_rules[rule_]() {
							goto l49
						}
						if buffer[position] != rune(',') {
							goto l49
						}
						position++
						if !_rules[rule_]() {
							goto l49
						}
						if !_rules[ruleValue]() {
							goto l49
						}
						if !_rules[rule_]() {
							goto l49
						}
						if buffer[position] != rune(']') {
							goto l49
						}
						position++
					}
				l60:
					if !_rules[ruleAction7]() {
						goto l49
					}
					goto l50
				l49:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
				}
			l50:
				depth--
				add(ruleBetweenExpression, position48)
			}
			return true
		l47:
			position, tokenIndex, depth = position47, tokenIndex47, depth47
			return false
		},
		/* 8 PrimaryExpression <- <(('(' _ Expression _ ')' Action8) / Value / (Identifier ArgumentList Action9) / Reference)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				{
					position70, tokenIndex70, depth70 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l71
					}
					position++
					if !_rules[rule_]() {
						goto l71
					}
					if !_rules[ruleExpression]() {
						goto l71
					}
					if !_rules[rule_]() {
						goto l71
					}
					if buffer[position] != rune(')') {
						goto l71
					}
					position++
					if !_rules[ruleAction8]() {
						goto l71
					}
					goto l70
				l71:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
					if !_rules[ruleValue]() {
						goto l72
					}
					goto l70
				l72:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
					if !_rules[ruleIdentifier]() {
						goto l73
					}
					if !_rules[ruleArgumentList]() {
						goto l73
					}
					if !_rules[ruleAction9]() {
						goto l73
					}
					goto l70
				l73:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
					if !_rules[ruleReference]() {
						goto l68
					}
				}
			l70:
				depth--
				add(rulePrimaryExpression, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 9 ArgumentList <- <('(' _ Action10 (Argument (_ ',' _ Argument)* _ ','?)? _ ')' Action11)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if buffer[position] != rune('(') {
					goto l74
				}
				position++
				if !_rules[rule_]() {
					goto l74
				}
				if !_rules[ruleAction10]() {
					goto l74
				}
				{
					position76, tokenIndex76, depth76 := position, tokenIndex, depth
					if !_rules[ruleArgument]() {
						goto l76
					}
				l78:
					{
						position79, tokenIndex79, depth79 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l79
						}
						if buffer[position] != rune(',') {
							goto l79
						}
						position++
						if !_rules[rule_]() {
							goto l79
						}
						if !_rules[ruleArgument]() {
							goto l79
						}
						goto l78
					l79:
						position, tokenIndex, depth = position79, tokenIndex79, depth79
					}
					if !_rules[rule_]() {
						goto l76
					}
					{
						position80, tokenIndex80, depth80 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l80
						}
						position++
						goto l81
					l80:
						position, tokenIndex, depth = position80, tokenIndex80, depth80
					}
				l81:
					goto l77
				l76:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
				}
			l77:
				if !_rules[rule_]() {
					goto l74
				}
				if buffer[position] != rune(')') {
					goto l74
				}
				position++
				if !_rules[ruleAction11]() {
					goto l74
				}
				depth--
				add(ruleArgumentList, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 10 Argument <- <Expression> */
		func() bool {
			position82, tokenIndex82, depth82 := position, tokenIndex, depth
			{
				position83 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l82
				}
				depth--
				add(ruleArgument, position83)
			}
			return true
		l82:
			position, tokenIndex, depth = position82, tokenIndex82, depth82
			return false
		},
		/* 11 Reference <- <((Identifier '.' Reference) / JsonReference / Identifier)> */
		func() bool {
			position84, tokenIndex84, depth84 := position, tokenIndex, depth
			{
				position85 := position
				depth++
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					if !_rules[ruleIdentifier]() {
						goto l87
					}
					if buffer[position] != rune('.') {
						goto l87
					}
					position++
					if !_rules[ruleReference]() {
						goto l87
					}
					goto l86
				l87:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if !_rules[ruleJsonReference]() {
						goto l88
					}
					goto l86
				l88:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if !_rules[ruleIdentifier]() {
						goto l84
					}
				}
			l86:
				depth--
				add(ruleReference, position85)
			}
			return true
		l84:
			position, tokenIndex, depth = position84, tokenIndex84, depth84
			return false
		},
		/* 12 JsonReference <- <(Identifier ('-' '>') (JsonReference / Identifier))> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if !_rules[ruleIdentifier]() {
					goto l89
				}
				if buffer[position] != rune('-') {
					goto l89
				}
				position++
				if buffer[position] != rune('>') {
					goto l89
				}
				position++
				{
					position91, tokenIndex91, depth91 := position, tokenIndex, depth
					if !_rules[ruleJsonReference]() {
						goto l92
					}
					goto l91
				l92:
					position, tokenIndex, depth = position91, tokenIndex91, depth91
					if !_rules[ruleIdentifier]() {
						goto l89
					}
				}
			l91:
				depth--
				add(ruleJsonReference, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 13 Identifier <- <(!(('n' / 'N') ('o' / 'O') ('t' / 'T')) <(([a-z] / [A-Z]) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('_') '_') | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> Action12)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					{
						position96, tokenIndex96, depth96 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l97
						}
						position++
						goto l96
					l97:
						position, tokenIndex, depth = position96, tokenIndex96, depth96
						if buffer[position] != rune('N') {
							goto l95
						}
						position++
					}
				l96:
					{
						position98, tokenIndex98, depth98 := position, tokenIndex, depth
						if buffer[position] != rune('o') {
							goto l99
						}
						position++
						goto l98
					l99:
						position, tokenIndex, depth = position98, tokenIndex98, depth98
						if buffer[position] != rune('O') {
							goto l95
						}
						position++
					}
				l98:
					{
						position100, tokenIndex100, depth100 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l101
						}
						position++
						goto l100
					l101:
						position, tokenIndex, depth = position100, tokenIndex100, depth100
						if buffer[position] != rune('T') {
							goto l95
						}
						position++
					}
				l100:
					goto l93
				l95:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
				}
				{
					position102 := position
					depth++
					{
						position103, tokenIndex103, depth103 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l104
						}
						position++
						goto l103
					l104:
						position, tokenIndex, depth = position103, tokenIndex103, depth103
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l93
						}
						position++
					}
				l103:
				l105:
					{
						position106, tokenIndex106, depth106 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l106
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l106
								}
								position++
								break
							case '_':
								if buffer[position] != rune('_') {
									goto l106
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l106
								}
								position++
								break
							}
						}

						goto l105
					l106:
						position, tokenIndex, depth = position106, tokenIndex106, depth106
					}
					depth--
					add(rulePegText, position102)
				}
				if !_rules[ruleAction12]() {
					goto l93
				}
				depth--
				add(ruleIdentifier, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 14 Compare <- <((_ <(('>' '=') / ('<' '=') / ('=' '=') / '<' / ((&('=') '=') | (&('<') ('<' '>')) | (&('>') '>') | (&('!') ('!' '='))))> _ Action13) / (_ <((('g' / 'G') ('t' / 'T')) / (('l' / 'L') ('t' / 'T')) / ((&('N' | 'n') (('n' / 'N') ('e' / 'E') ('q' / 'Q'))) | (&('E' | 'e') (('e' / 'E') ('q' / 'Q'))) | (&('L' | 'l') (('l' / 'L') ('t' / 'T') ('e' / 'E'))) | (&('G' | 'g') (('g' / 'G') ('t' / 'T') ('e' / 'E')))))> _ Action14) / (_ <((('l' / 'L') ('i' / 'I') ('k' / 'K') ('e' / 'E')) / (('n' / 'N') ('o' / 'O') ('t' / 'T') __ (('l' / 'L') ('i' / 'I') ('k' / 'K') ('e' / 'E'))))> _ Action15))> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l111
					}
					{
						position112 := position
						depth++
						{
							position113, tokenIndex113, depth113 := position, tokenIndex, depth
							if buffer[position] != rune('>') {
								goto l114
							}
							position++
							if buffer[position] != rune('=') {
								goto l114
							}
							position++
							goto l113
						l114:
							position, tokenIndex, depth = position113, tokenIndex113, depth113
							if buffer[position] != rune('<') {
								goto l115
							}
							position++
							if buffer[position] != rune('=') {
								goto l115
							}
							position++
							goto l113
						l115:
							position, tokenIndex, depth = position113, tokenIndex113, depth113
							if buffer[position] != rune('=') {
								goto l116
							}
							position++
							if buffer[position] != rune('=') {
								goto l116
							}
							position++
							goto l113
						l116:
							position, tokenIndex, depth = position113, tokenIndex113, depth113
							if buffer[position] != rune('<') {
								goto l117
							}
							position++
							goto l113
						l117:
							position, tokenIndex, depth = position113, tokenIndex113, depth113
							{
								switch buffer[position] {
								case '=':
									if buffer[position] != rune('=') {
										goto l111
									}
									position++
									break
								case '<':
									if buffer[position] != rune('<') {
										goto l111
									}
									position++
									if buffer[position] != rune('>') {
										goto l111
									}
									position++
									break
								case '>':
									if buffer[position] != rune('>') {
										goto l111
									}
									position++
									break
								default:
									if buffer[position] != rune('!') {
										goto l111
									}
									position++
									if buffer[position] != rune('=') {
										goto l111
									}
									position++
									break
								}
							}

						}
					l113:
						depth--
						add(rulePegText, position112)
					}
					if !_rules[rule_]() {
						goto l111
					}
					if !_rules[ruleAction13]() {
						goto l111
					}
					goto l110
				l111:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[rule_]() {
						goto l119
					}
					{
						position120 := position
						depth++
						{
							position121, tokenIndex121, depth121 := position, tokenIndex, depth
							{
								position123, tokenIndex123, depth123 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l124
								}
								position++
								goto l123
							l124:
								position, tokenIndex, depth = position123, tokenIndex123, depth123
								if buffer[position] != rune('G') {
									goto l122
								}
								position++
							}
						l123:
							{
								position125, tokenIndex125, depth125 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l126
								}
								position++
								goto l125
							l126:
								position, tokenIndex, depth = position125, tokenIndex125, depth125
								if buffer[position] != rune('T') {
									goto l122
								}
								position++
							}
						l125:
							goto l121
						l122:
							position, tokenIndex, depth = position121, tokenIndex121, depth121
							{
								position128, tokenIndex128, depth128 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l129
								}
								position++
								goto l128
							l129:
								position, tokenIndex, depth = position128, tokenIndex128, depth128
								if buffer[position] != rune('L') {
									goto l127
								}
								position++
							}
						l128:
							{
								position130, tokenIndex130, depth130 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l131
								}
								position++
								goto l130
							l131:
								position, tokenIndex, depth = position130, tokenIndex130, depth130
								if buffer[position] != rune('T') {
									goto l127
								}
								position++
							}
						l130:
							goto l121
						l127:
							position, tokenIndex, depth = position121, tokenIndex121, depth121
							{
								switch buffer[position] {
								case 'N', 'n':
									{
										position133, tokenIndex133, depth133 := position, tokenIndex, depth
										if buffer[position] != rune('n') {
											goto l134
										}
										position++
										goto l133
									l134:
										position, tokenIndex, depth = position133, tokenIndex133, depth133
										if buffer[position] != rune('N') {
											goto l119
										}
										position++
									}
								l133:
									{
										position135, tokenIndex135, depth135 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l136
										}
										position++
										goto l135
									l136:
										position, tokenIndex, depth = position135, tokenIndex135, depth135
										if buffer[position] != rune('E') {
											goto l119
										}
										position++
									}
								l135:
									{
										position137, tokenIndex137, depth137 := position, tokenIndex, depth
										if buffer[position] != rune('q') {
											goto l138
										}
										position++
										goto l137
									l138:
										position, tokenIndex, depth = position137, tokenIndex137, depth137
										if buffer[position] != rune('Q') {
											goto l119
										}
										position++
									}
								l137:
									break
								case 'E', 'e':
									{
										position139, tokenIndex139, depth139 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l140
										}
										position++
										goto l139
									l140:
										position, tokenIndex, depth = position139, tokenIndex139, depth139
										if buffer[position] != rune('E') {
											goto l119
										}
										position++
									}
								l139:
									{
										position141, tokenIndex141, depth141 := position, tokenIndex, depth
										if buffer[position] != rune('q') {
											goto l142
										}
										position++
										goto l141
									l142:
										position, tokenIndex, depth = position141, tokenIndex141, depth141
										if buffer[position] != rune('Q') {
											goto l119
										}
										position++
									}
								l141:
									break
								case 'L', 'l':
									{
										position143, tokenIndex143, depth143 := position, tokenIndex, depth
										if buffer[position] != rune('l') {
											goto l144
										}
										position++
										goto l143
									l144:
										position, tokenIndex, depth = position143, tokenIndex143, depth143
										if buffer[position] != rune('L') {
											goto l119
										}
										position++
									}
								l143:
									{
										position145, tokenIndex145, depth145 := position, tokenIndex, depth
										if buffer[position] != rune('t') {
											goto l146
										}
										position++
										goto l145
									l146:
										position, tokenIndex, depth = position145, tokenIndex145, depth145
										if buffer[position] != rune('T') {
											goto l119
										}
										position++
									}
								l145:
									{
										position147, tokenIndex147, depth147 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l148
										}
										position++
										goto l147
									l148:
										position, tokenIndex, depth = position147, tokenIndex147, depth147
										if buffer[position] != rune('E') {
											goto l119
										}
										position++
									}
								l147:
									break
								default:
									{
										position149, tokenIndex149, depth149 := position, tokenIndex, depth
										if buffer[position] != rune('g') {
											goto l150
										}
										position++
										goto l149
									l150:
										position, tokenIndex, depth = position149, tokenIndex149, depth149
										if buffer[position] != rune('G') {
											goto l119
										}
										position++
									}
								l149:
									{
										position151, tokenIndex151, depth151 := position, tokenIndex, depth
										if buffer[position] != rune('t') {
											goto l152
										}
										position++
										goto l151
									l152:
										position, tokenIndex, depth = position151, tokenIndex151, depth151
										if buffer[position] != rune('T') {
											goto l119
										}
										position++
									}
								l151:
									{
										position153, tokenIndex153, depth153 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l154
										}
										position++
										goto l153
									l154:
										position, tokenIndex, depth = position153, tokenIndex153, depth153
										if buffer[position] != rune('E') {
											goto l119
										}
										position++
									}
								l153:
									break
								}
							}

						}
					l121:
						depth--
						add(rulePegText, position120)
					}
					if !_rules[rule_]() {
						goto l119
					}
					if !_rules[ruleAction14]() {
						goto l119
					}
					goto l110
				l119:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[rule_]() {
						goto l108
					}
					{
						position155 := position
						depth++
						{
							position156, tokenIndex156, depth156 := position, tokenIndex, depth
							{
								position158, tokenIndex158, depth158 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l159
								}
								position++
								goto l158
							l159:
								position, tokenIndex, depth = position158, tokenIndex158, depth158
								if buffer[position] != rune('L') {
									goto l157
								}
								position++
							}
						l158:
							{
								position160, tokenIndex160, depth160 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l161
								}
								position++
								goto l160
							l161:
								position, tokenIndex, depth = position160, tokenIndex160, depth160
								if buffer[position] != rune('I') {
									goto l157
								}
								position++
							}
						l160:
							{
								position162, tokenIndex162, depth162 := position, tokenIndex, depth
								if buffer[position] != rune('k') {
									goto l163
								}
								position++
								goto l162
							l163:
								position, tokenIndex, depth = position162, tokenIndex162, depth162
								if buffer[position] != rune('K') {
									goto l157
								}
								position++
							}
						l162:
							{
								position164, tokenIndex164, depth164 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l165
								}
								position++
								goto l164
							l165:
								position, tokenIndex, depth = position164, tokenIndex164, depth164
								if buffer[position] != rune('E') {
									goto l157
								}
								position++
							}
						l164:
							goto l156
						l157:
							position, tokenIndex, depth = position156, tokenIndex156, depth156
							{
								position166, tokenIndex166, depth166 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l167
								}
								position++
								goto l166
							l167:
								position, tokenIndex, depth = position166, tokenIndex166, depth166
								if buffer[position] != rune('N') {
									goto l108
								}
								position++
							}
						l166:
							{
								position168, tokenIndex168, depth168 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l169
								}
								position++
								goto l168
							l169:
								position, tokenIndex, depth = position168, tokenIndex168, depth168
								if buffer[position] != rune('O') {
									goto l108
								}
								position++
							}
						l168:
							{
								position170, tokenIndex170, depth170 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l171
								}
								position++
								goto l170
							l171:
								position, tokenIndex, depth = position170, tokenIndex170, depth170
								if buffer[position] != rune('T') {
									goto l108
								}
								position++
							}
						l170:
							if !_rules[rule__]() {
								goto l108
							}
							{
								position172, tokenIndex172, depth172 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l173
								}
								position++
								goto l172
							l173:
								position, tokenIndex, depth = position172, tokenIndex172, depth172
								if buffer[position] != rune('L') {
									goto l108
								}
								position++
							}
						l172:
							{
								position174, tokenIndex174, depth174 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l175
								}
								position++
								goto l174
							l175:
								position, tokenIndex, depth = position174, tokenIndex174, depth174
								if buffer[position] != rune('I') {
									goto l108
								}
								position++
							}
						l174:
							{
								position176, tokenIndex176, depth176 := position, tokenIndex, depth
								if buffer[position] != rune('k') {
									goto l177
								}
								position++
								goto l176
							l177:
								position, tokenIndex, depth = position176, tokenIndex176, depth176
								if buffer[position] != rune('K') {
									goto l108
								}
								position++
							}
						l176:
							{
								position178, tokenIndex178, depth178 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l179
								}
								position++
								goto l178
							l179:
								position, tokenIndex, depth = position178, tokenIndex178, depth178
								if buffer[position] != rune('E') {
									goto l108
								}
								position++
							}
						l178:
						}
					l156:
						depth--
						add(rulePegText, position155)
					}
					if !_rules[rule_]() {
						goto l108
					}
					if !_rules[ruleAction15]() {
						goto l108
					}
				}
			l110:
				depth--
				add(ruleCompare, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 15 Logic <- <((_ <((('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('o' / 'O') ('r' / 'R')))> _ Action16) / (_ <(('&' '&') / ('|' '|'))> _ Action17))> */
		func() bool {
			position180, tokenIndex180, depth180 := position, tokenIndex, depth
			{
				position181 := position
				depth++
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l183
					}
					{
						position184 := position
						depth++
						{
							position185, tokenIndex185, depth185 := position, tokenIndex, depth
							{
								position187, tokenIndex187, depth187 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l188
								}
								position++
								goto l187
							l188:
								position, tokenIndex, depth = position187, tokenIndex187, depth187
								if buffer[position] != rune('A') {
									goto l186
								}
								position++
							}
						l187:
							{
								position189, tokenIndex189, depth189 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l190
								}
								position++
								goto l189
							l190:
								position, tokenIndex, depth = position189, tokenIndex189, depth189
								if buffer[position] != rune('N') {
									goto l186
								}
								position++
							}
						l189:
							{
								position191, tokenIndex191, depth191 := position, tokenIndex, depth
								if buffer[position] != rune('d') {
									goto l192
								}
								position++
								goto l191
							l192:
								position, tokenIndex, depth = position191, tokenIndex191, depth191
								if buffer[position] != rune('D') {
									goto l186
								}
								position++
							}
						l191:
							goto l185
						l186:
							position, tokenIndex, depth = position185, tokenIndex185, depth185
							{
								position193, tokenIndex193, depth193 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l194
								}
								position++
								goto l193
							l194:
								position, tokenIndex, depth = position193, tokenIndex193, depth193
								if buffer[position] != rune('O') {
									goto l183
								}
								position++
							}
						l193:
							{
								position195, tokenIndex195, depth195 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l196
								}
								position++
								goto l195
							l196:
								position, tokenIndex, depth = position195, tokenIndex195, depth195
								if buffer[position] != rune('R') {
									goto l183
								}
								position++
							}
						l195:
						}
					l185:
						depth--
						add(rulePegText, position184)
					}
					if !_rules[rule_]() {
						goto l183
					}
					if !_rules[ruleAction16]() {
						goto l183
					}
					goto l182
				l183:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
					if !_rules[rule_]() {
						goto l180
					}
					{
						position197 := position
						depth++
						{
							position198, tokenIndex198, depth198 := position, tokenIndex, depth
							if buffer[position] != rune('&') {
								goto l199
							}
							position++
							if buffer[position] != rune('&') {
								goto l199
							}
							position++
							goto l198
						l199:
							position, tokenIndex, depth = position198, tokenIndex198, depth198
							if buffer[position] != rune('|') {
								goto l180
							}
							position++
							if buffer[position] != rune('|') {
								goto l180
							}
							position++
						}
					l198:
						depth--
						add(rulePegText, position197)
					}
					if !_rules[rule_]() {
						goto l180
					}
					if !_rules[ruleAction17]() {
						goto l180
					}
				}
			l182:
				depth--
				add(ruleLogic, position181)
			}
			return true
		l180:
			position, tokenIndex, depth = position180, tokenIndex180, depth180
			return false
		},
		/* 16 Match <- <(__ <(('i' 's' 'n' 'u' 'l' 'l') / ('n' 'o' 't' 'n' 'u' 'l' 'l') / ('i' 's' __ ((&('n') ('n' 'u' 'l' 'l')) | (&('f') ('f' 'a' 'l' 's' 'e')) | (&('t') ('t' 'r' 'u' 'e')))) / ('i' 's' __ ('n' 'o' 't') __ ((&('n') ('n' 'u' 'l' 'l')) | (&('f') ('f' 'a' 'l' 's' 'e')) | (&('t') ('t' 'r' 'u' 'e')))))> _ Action18)> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				if !_rules[rule__]() {
					goto l200
				}
				{
					position202 := position
					depth++
					{
						position203, tokenIndex203, depth203 := position, tokenIndex, depth
						if buffer[position] != rune('i') {
							goto l204
						}
						position++
						if buffer[position] != rune('s') {
							goto l204
						}
						position++
						if buffer[position] != rune('n') {
							goto l204
						}
						position++
						if buffer[position] != rune('u') {
							goto l204
						}
						position++
						if buffer[position] != rune('l') {
							goto l204
						}
						position++
						if buffer[position] != rune('l') {
							goto l204
						}
						position++
						goto l203
					l204:
						position, tokenIndex, depth = position203, tokenIndex203, depth203
						if buffer[position] != rune('n') {
							goto l205
						}
						position++
						if buffer[position] != rune('o') {
							goto l205
						}
						position++
						if buffer[position] != rune('t') {
							goto l205
						}
						position++
						if buffer[position] != rune('n') {
							goto l205
						}
						position++
						if buffer[position] != rune('u') {
							goto l205
						}
						position++
						if buffer[position] != rune('l') {
							goto l205
						}
						position++
						if buffer[position] != rune('l') {
							goto l205
						}
						position++
						goto l203
					l205:
						position, tokenIndex, depth = position203, tokenIndex203, depth203
						if buffer[position] != rune('i') {
							goto l206
						}
						position++
						if buffer[position] != rune('s') {
							goto l206
						}
						position++
						if !_rules[rule__]() {
							goto l206
						}
						{
							switch buffer[position] {
							case 'n':
								if buffer[position] != rune('n') {
									goto l206
								}
								position++
								if buffer[position] != rune('u') {
									goto l206
								}
								position++
								if buffer[position] != rune('l') {
									goto l206
								}
								position++
								if buffer[position] != rune('l') {
									goto l206
								}
								position++
								break
							case 'f':
								if buffer[position] != rune('f') {
									goto l206
								}
								position++
								if buffer[position] != rune('a') {
									goto l206
								}
								position++
								if buffer[position] != rune('l') {
									goto l206
								}
								position++
								if buffer[position] != rune('s') {
									goto l206
								}
								position++
								if buffer[position] != rune('e') {
									goto l206
								}
								position++
								break
							default:
								if buffer[position] != rune('t') {
									goto l206
								}
								position++
								if buffer[position] != rune('r') {
									goto l206
								}
								position++
								if buffer[position] != rune('u') {
									goto l206
								}
								position++
								if buffer[position] != rune('e') {
									goto l206
								}
								position++
								break
							}
						}

						goto l203
					l206:
						position, tokenIndex, depth = position203, tokenIndex203, depth203
						if buffer[position] != rune('i') {
							goto l200
						}
						position++
						if buffer[position] != rune('s') {
							goto l200
						}
						position++
						if !_rules[rule__]() {
							goto l200
						}
						if buffer[position] != rune('n') {
							goto l200
						}
						position++
						if buffer[position] != rune('o') {
							goto l200
						}
						position++
						if buffer[position] != rune('t') {
							goto l200
						}
						position++
						if !_rules[rule__]() {
							goto l200
						}
						{
							switch buffer[position] {
							case 'n':
								if buffer[position] != rune('n') {
									goto l200
								}
								position++
								if buffer[position] != rune('u') {
									goto l200
								}
								position++
								if buffer[position] != rune('l') {
									goto l200
								}
								position++
								if buffer[position] != rune('l') {
									goto l200
								}
								position++
								break
							case 'f':
								if buffer[position] != rune('f') {
									goto l200
								}
								position++
								if buffer[position] != rune('a') {
									goto l200
								}
								position++
								if buffer[position] != rune('l') {
									goto l200
								}
								position++
								if buffer[position] != rune('s') {
									goto l200
								}
								position++
								if buffer[position] != rune('e') {
									goto l200
								}
								position++
								break
							default:
								if buffer[position] != rune('t') {
									goto l200
								}
								position++
								if buffer[position] != rune('r') {
									goto l200
								}
								position++
								if buffer[position] != rune('u') {
									goto l200
								}
								position++
								if buffer[position] != rune('e') {
									goto l200
								}
								position++
								break
							}
						}

					}
				l203:
					depth--
					add(rulePegText, position202)
				}
				if !_rules[rule_]() {
					goto l200
				}
				if !_rules[ruleAction18]() {
					goto l200
				}
				depth--
				add(ruleMatch, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 17 Value <- <(Literal / Array)> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if !_rules[ruleLiteral]() {
						goto l212
					}
					goto l211
				l212:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
					if !_rules[ruleArray]() {
						goto l209
					}
				}
			l211:
				depth--
				add(ruleValue, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 18 Array <- <(('[' Action19 _ (Literal (_ ',' _ Literal)* _ ','?)? _ ']' Action20) / ('(' Action21 _ (Literal (_ ',' _ Literal)* _ ','?)? _ ')' Action22))> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l216
					}
					position++
					if !_rules[ruleAction19]() {
						goto l216
					}
					if !_rules[rule_]() {
						goto l216
					}
					{
						position217, tokenIndex217, depth217 := position, tokenIndex, depth
						if !_rules[ruleLiteral]() {
							goto l217
						}
					l219:
						{
							position220, tokenIndex220, depth220 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l220
							}
							if buffer[position] != rune(',') {
								goto l220
							}
							position++
							if !_rules[rule_]() {
								goto l220
							}
							if !_rules[ruleLiteral]() {
								goto l220
							}
							goto l219
						l220:
							position, tokenIndex, depth = position220, tokenIndex220, depth220
						}
						if !_rules[rule_]() {
							goto l217
						}
						{
							position221, tokenIndex221, depth221 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l221
							}
							position++
							goto l222
						l221:
							position, tokenIndex, depth = position221, tokenIndex221, depth221
						}
					l222:
						goto l218
					l217:
						position, tokenIndex, depth = position217, tokenIndex217, depth217
					}
				l218:
					if !_rules[rule_]() {
						goto l216
					}
					if buffer[position] != rune(']') {
						goto l216
					}
					position++
					if !_rules[ruleAction20]() {
						goto l216
					}
					goto l215
				l216:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
					if buffer[position] != rune('(') {
						goto l213
					}
					position++
					if !_rules[ruleAction21]() {
						goto l213
					}
					if !_rules[rule_]() {
						goto l213
					}
					{
						position223, tokenIndex223, depth223 := position, tokenIndex, depth
						if !_rules[ruleLiteral]() {
							goto l223
						}
					l225:
						{
							position226, tokenIndex226, depth226 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l226
							}
							if buffer[position] != rune(',') {
								goto l226
							}
							position++
							if !_rules[rule_]() {
								goto l226
							}
							if !_rules[ruleLiteral]() {
								goto l226
							}
							goto l225
						l226:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
						}
						if !_rules[rule_]() {
							goto l223
						}
						{
							position227, tokenIndex227, depth227 := position, tokenIndex, depth
							if buffer[position] != rune(',') {
								goto l227
							}
							position++
							goto l228
						l227:
							position, tokenIndex, depth = position227, tokenIndex227, depth227
						}
					l228:
						goto l224
					l223:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
					}
				l224:
					if !_rules[rule_]() {
						goto l213
					}
					if buffer[position] != rune(')') {
						goto l213
					}
					position++
					if !_rules[ruleAction22]() {
						goto l213
					}
				}
			l215:
				depth--
				add(ruleArray, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 19 Literal <- <((&('N' | 'n') Null) | (&('F' | 'T' | 'f' | 't') Boolean) | (&('"' | '\'') String) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') Integer))> */
		func() bool {
			position229, tokenIndex229, depth229 := position, tokenIndex, depth
			{
				position230 := position
				depth++
				{
					switch buffer[position] {
					case 'N', 'n':
						if !_rules[ruleNull]() {
							goto l229
						}
						break
					case 'F', 'T', 'f', 't':
						if !_rules[ruleBoolean]() {
							goto l229
						}
						break
					case '"', '\'':
						if !_rules[ruleString]() {
							goto l229
						}
						break
					default:
						if !_rules[ruleInteger]() {
							goto l229
						}
						break
					}
				}

				depth--
				add(ruleLiteral, position230)
			}
			return true
		l229:
			position, tokenIndex, depth = position229, tokenIndex229, depth229
			return false
		},
		/* 20 Integer <- <(<('0' / ([1-9] [0-9]*))> Action23)> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				{
					position234 := position
					depth++
					{
						position235, tokenIndex235, depth235 := position, tokenIndex, depth
						if buffer[position] != rune('0') {
							goto l236
						}
						position++
						goto l235
					l236:
						position, tokenIndex, depth = position235, tokenIndex235, depth235
						if c := buffer[position]; c < rune('1') || c > rune('9') {
							goto l232
						}
						position++
					l237:
						{
							position238, tokenIndex238, depth238 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l238
							}
							position++
							goto l237
						l238:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
						}
					}
				l235:
					depth--
					add(rulePegText, position234)
				}
				if !_rules[ruleAction23]() {
					goto l232
				}
				depth--
				add(ruleInteger, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 21 Boolean <- <(<((&('F') ('F' 'A' 'L' 'S' 'E')) | (&('T') ('T' 'R' 'U' 'E')) | (&('f') ('f' 'a' 'l' 's' 'e')) | (&('t') ('t' 'r' 'u' 'e')))> Action24)> */
		func() bool {
			position239, tokenIndex239, depth239 := position, tokenIndex, depth
			{
				position240 := position
				depth++
				{
					position241 := position
					depth++
					{
						switch buffer[position] {
						case 'F':
							if buffer[position] != rune('F') {
								goto l239
							}
							position++
							if buffer[position] != rune('A') {
								goto l239
							}
							position++
							if buffer[position] != rune('L') {
								goto l239
							}
							position++
							if buffer[position] != rune('S') {
								goto l239
							}
							position++
							if buffer[position] != rune('E') {
								goto l239
							}
							position++
							break
						case 'T':
							if buffer[position] != rune('T') {
								goto l239
							}
							position++
							if buffer[position] != rune('R') {
								goto l239
							}
							position++
							if buffer[position] != rune('U') {
								goto l239
							}
							position++
							if buffer[position] != rune('E') {
								goto l239
							}
							position++
							break
						case 'f':
							if buffer[position] != rune('f') {
								goto l239
							}
							position++
							if buffer[position] != rune('a') {
								goto l239
							}
							position++
							if buffer[position] != rune('l') {
								goto l239
							}
							position++
							if buffer[position] != rune('s') {
								goto l239
							}
							position++
							if buffer[position] != rune('e') {
								goto l239
							}
							position++
							break
						default:
							if buffer[position] != rune('t') {
								goto l239
							}
							position++
							if buffer[position] != rune('r') {
								goto l239
							}
							position++
							if buffer[position] != rune('u') {
								goto l239
							}
							position++
							if buffer[position] != rune('e') {
								goto l239
							}
							position++
							break
						}
					}

					depth--
					add(rulePegText, position241)
				}
				if !_rules[ruleAction24]() {
					goto l239
				}
				depth--
				add(ruleBoolean, position240)
			}
			return true
		l239:
			position, tokenIndex, depth = position239, tokenIndex239, depth239
			return false
		},
		/* 22 Null <- <(<(('n' 'u' 'l' 'l') / ('N' 'U' 'L' 'L'))> Action25)> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				{
					position245 := position
					depth++
					{
						position246, tokenIndex246, depth246 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l247
						}
						position++
						if buffer[position] != rune('u') {
							goto l247
						}
						position++
						if buffer[position] != rune('l') {
							goto l247
						}
						position++
						if buffer[position] != rune('l') {
							goto l247
						}
						position++
						goto l246
					l247:
						position, tokenIndex, depth = position246, tokenIndex246, depth246
						if buffer[position] != rune('N') {
							goto l243
						}
						position++
						if buffer[position] != rune('U') {
							goto l243
						}
						position++
						if buffer[position] != rune('L') {
							goto l243
						}
						position++
						if buffer[position] != rune('L') {
							goto l243
						}
						position++
					}
				l246:
					depth--
					add(rulePegText, position245)
				}
				if !_rules[ruleAction25]() {
					goto l243
				}
				depth--
				add(ruleNull, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 23 String <- <(('\'' <(!'\'' .)*> '\'' Action26) / ('"' <(!'"' .)*> '"' Action27))> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l251
					}
					position++
					{
						position252 := position
						depth++
					l253:
						{
							position254, tokenIndex254, depth254 := position, tokenIndex, depth
							{
								position255, tokenIndex255, depth255 := position, tokenIndex, depth
								if buffer[position] != rune('\'') {
									goto l255
								}
								position++
								goto l254
							l255:
								position, tokenIndex, depth = position255, tokenIndex255, depth255
							}
							if !matchDot() {
								goto l254
							}
							goto l253
						l254:
							position, tokenIndex, depth = position254, tokenIndex254, depth254
						}
						depth--
						add(rulePegText, position252)
					}
					if buffer[position] != rune('\'') {
						goto l251
					}
					position++
					if !_rules[ruleAction26]() {
						goto l251
					}
					goto l250
				l251:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
					if buffer[position] != rune('"') {
						goto l248
					}
					position++
					{
						position256 := position
						depth++
					l257:
						{
							position258, tokenIndex258, depth258 := position, tokenIndex, depth
							{
								position259, tokenIndex259, depth259 := position, tokenIndex, depth
								if buffer[position] != rune('"') {
									goto l259
								}
								position++
								goto l258
							l259:
								position, tokenIndex, depth = position259, tokenIndex259, depth259
							}
							if !matchDot() {
								goto l258
							}
							goto l257
						l258:
							position, tokenIndex, depth = position258, tokenIndex258, depth258
						}
						depth--
						add(rulePegText, position256)
					}
					if buffer[position] != rune('"') {
						goto l248
					}
					position++
					if !_rules[ruleAction27]() {
						goto l248
					}
				}
			l250:
				depth--
				add(ruleString, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 24 SpaceComment <- <(Space / Comment)> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				{
					position262, tokenIndex262, depth262 := position, tokenIndex, depth
					if !_rules[ruleSpace]() {
						goto l263
					}
					goto l262
				l263:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
					if !_rules[ruleComment]() {
						goto l260
					}
				}
			l262:
				depth--
				add(ruleSpaceComment, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 25 _ <- <SpaceComment*> */
		func() bool {
			{
				position265 := position
				depth++
			l266:
				{
					position267, tokenIndex267, depth267 := position, tokenIndex, depth
					if !_rules[ruleSpaceComment]() {
						goto l267
					}
					goto l266
				l267:
					position, tokenIndex, depth = position267, tokenIndex267, depth267
				}
				depth--
				add(rule_, position265)
			}
			return true
		},
		/* 26 __ <- <SpaceComment+> */
		func() bool {
			position268, tokenIndex268, depth268 := position, tokenIndex, depth
			{
				position269 := position
				depth++
				if !_rules[ruleSpaceComment]() {
					goto l268
				}
			l270:
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					if !_rules[ruleSpaceComment]() {
						goto l271
					}
					goto l270
				l271:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
				}
				depth--
				add(rule__, position269)
			}
			return true
		l268:
			position, tokenIndex, depth = position268, tokenIndex268, depth268
			return false
		},
		/* 27 Comment <- <((('-' '-') / ('/' '/')) (!EndOfLine .)* EndOfLine)> */
		func() bool {
			position272, tokenIndex272, depth272 := position, tokenIndex, depth
			{
				position273 := position
				depth++
				{
					position274, tokenIndex274, depth274 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l275
					}
					position++
					if buffer[position] != rune('-') {
						goto l275
					}
					position++
					goto l274
				l275:
					position, tokenIndex, depth = position274, tokenIndex274, depth274
					if buffer[position] != rune('/') {
						goto l272
					}
					position++
					if buffer[position] != rune('/') {
						goto l272
					}
					position++
				}
			l274:
			l276:
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					{
						position278, tokenIndex278, depth278 := position, tokenIndex, depth
						if !_rules[ruleEndOfLine]() {
							goto l278
						}
						goto l277
					l278:
						position, tokenIndex, depth = position278, tokenIndex278, depth278
					}
					if !matchDot() {
						goto l277
					}
					goto l276
				l277:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
				}
				if !_rules[ruleEndOfLine]() {
					goto l272
				}
				depth--
				add(ruleComment, position273)
			}
			return true
		l272:
			position, tokenIndex, depth = position272, tokenIndex272, depth272
			return false
		},
		/* 28 Space <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') EndOfLine))> */
		func() bool {
			position279, tokenIndex279, depth279 := position, tokenIndex, depth
			{
				position280 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l279
						}
						position++
						break
					case ' ':
						if buffer[position] != rune(' ') {
							goto l279
						}
						position++
						break
					default:
						if !_rules[ruleEndOfLine]() {
							goto l279
						}
						break
					}
				}

				depth--
				add(ruleSpace, position280)
			}
			return true
		l279:
			position, tokenIndex, depth = position279, tokenIndex279, depth279
			return false
		},
		/* 29 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position282, tokenIndex282, depth282 := position, tokenIndex, depth
			{
				position283 := position
				depth++
				{
					position284, tokenIndex284, depth284 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l285
					}
					position++
					if buffer[position] != rune('\n') {
						goto l285
					}
					position++
					goto l284
				l285:
					position, tokenIndex, depth = position284, tokenIndex284, depth284
					if buffer[position] != rune('\n') {
						goto l286
					}
					position++
					goto l284
				l286:
					position, tokenIndex, depth = position284, tokenIndex284, depth284
					if buffer[position] != rune('\r') {
						goto l282
					}
					position++
				}
			l284:
				depth--
				add(ruleEndOfLine, position283)
			}
			return true
		l282:
			position, tokenIndex, depth = position282, tokenIndex282, depth282
			return false
		},
		/* 30 EndOfFile <- <!.> */
		func() bool {
			position287, tokenIndex287, depth287 := position, tokenIndex, depth
			{
				position288 := position
				depth++
				{
					position289, tokenIndex289, depth289 := position, tokenIndex, depth
					if !matchDot() {
						goto l289
					}
					goto l287
				l289:
					position, tokenIndex, depth = position289, tokenIndex289, depth289
				}
				depth--
				add(ruleEndOfFile, position288)
			}
			return true
		l287:
			position, tokenIndex, depth = position287, tokenIndex287, depth287
			return false
		},
		/* 32 Action0 <- <{p.PopLogic()}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 33 Action1 <- <{p.PopNot()}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 34 Action2 <- <{p.PopCompare()}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		nil,
		/* 36 Action3 <- <{p.AddCompare(text)}> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 37 Action4 <- <{p.PopCompare()}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 38 Action5 <- <{p.PopPredicate()}> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 39 Action6 <- <{p.AddOperation(text)}> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 40 Action7 <- <{p.PopBetween()}> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 41 Action8 <- <{p.PopParentheses()}> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 42 Action9 <- <{p.PopFunction()}> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 43 Action10 <- <{p.AddMark()}> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 44 Action11 <- <{p.PopArray()}> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 45 Action12 <- <{p.AddName(text)}> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 46 Action13 <- <{p.AddCompare(text)}> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 47 Action14 <- <{p.AddCompare(text)}> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 48 Action15 <- <{p.AddCompare(text)}> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 49 Action16 <- <{p.AddLogic(text)}> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		/* 50 Action17 <- <{p.AddLogic(text)}> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 51 Action18 <- <{p.AddMatch(text)}> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 52 Action19 <- <{p.AddMark()}> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 53 Action20 <- <{p.PopArray()}> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 54 Action21 <- <{p.AddMark()}> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 55 Action22 <- <{p.PopArray()}> */
		func() bool {
			{
				add(ruleAction22, position)
			}
			return true
		},
		/* 56 Action23 <- <{p.AddInteger(text)}> */
		func() bool {
			{
				add(ruleAction23, position)
			}
			return true
		},
		/* 57 Action24 <- <{p.AddBoolean(text)}> */
		func() bool {
			{
				add(ruleAction24, position)
			}
			return true
		},
		/* 58 Action25 <- <{p.AddNull()}> */
		func() bool {
			{
				add(ruleAction25, position)
			}
			return true
		},
		/* 59 Action26 <- <{p.AddString(text)}> */
		func() bool {
			{
				add(ruleAction26, position)
			}
			return true
		},
		/* 60 Action27 <- <{p.AddString(text)}> */
		func() bool {
			{
				add(ruleAction27, position)
			}
			return true
		},
	}
	p.rules = _rules
}

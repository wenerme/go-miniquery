package miniquery

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/stretchr/testify/assert"
)

func TestParseFunc(t *testing.T) {
	n, err := Parse("1")
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Value())
}

func TestRareCase(t *testing.T) {
	for _, v := range []struct {
		E string
		Q string
	}{
		{E: "a in [1]", Q: "a in (1)"},
	} {
		p := &MiniQueryPeg{Tree: &Tree{}, Buffer: v.Q}
		assert.NoError(t, p.Init())
		if assert.NoError(t, p.Parse(), v) {
			p.Execute()
			fmt.Printf("%s\t---->\t %s : %v\n", v, Build(p.Stack[0]), p.Stack[0])
			assert.Equal(t, 0, len(p.Errors))
			if !assert.Equal(t, 1, len(p.Stack)) {
				spew.Dump(p.Stack)
			}
			assert.Equal(t, v.E, Build(p.Stack[0]))
		}
	}
}

func TestParsing(t *testing.T) {
	for _, v := range []string{
		"1=1 and a >= 1",
		"1=1 AND a >= 1",
		` ( a like '%hello%' ) `,
		`1=1 or 1=2 or 1=3 `,
		`(1=1) or( 1=2 )or( 1=3)or(1=3 ) `,
		`a=b or a=0`,
		`a is  not    null`,
		`a is      null and a = true`,
		`1 in [0,1,2,3,4] and b >= 2 or c not in []`,
		`1 in [ 0 , 1 , 2 , 3 , 4 ]`,
		`a in ("a", "b"  ,"c")`,
		`a in (1)`,
		`a between 12 and 15 or b not between 67 and 78`,
		`not true`,
		`not true or not 1 > 2`,
		`a is  true and b is  not false`,
		`owned = true`,
		`a between 1 and 2`,
		`a between [1,2]`, // 特殊的 between 语法，方便构造语法
		// 函数
		`a()`,
		`func(123)`,
		`func(1,2,3,4,5,)`,
		`func(0,true,false,null,'a',"bc")`,
		`func(['a',1,3])`,
		`date(a) = func(123)`,
		`null == NULL`,
		`func(a)`,
		`func(a,b)`,
		`func( a , b )`,
		`func(a,name==1)`,
		`func( a, name == 1 )`,
		`func( a, name in (1,2,3) )`,
		`func( a, name in ('a','b') )`,
		`date('2021-05-12T00:00:00+08:00')`,
		`profile.age > 10`,
		// unsupported
		// `date(created_at) between date('2021-05-12T00:00:00+08:00') and date('2021-05-14T00:00:00+08:00')`,
	} {
		p := &MiniQueryPeg{Tree: &Tree{}, Buffer: v}
		assert.NoError(t, p.Init())
		if !assert.NoError(t, p.Parse(), v) {
			fmt.Println("PrintSyntaxTree")
			p.PrintSyntaxTree()
			p.Execute()
			spew.Dump(p.Stack)
			fmt.Println("Marks", p.marks)
			fmt.Println("Errors", p.Errors)
		} else {
			p.Execute()
			fmt.Printf("%s\t---->\t %s : %v\n", v, Build(p.Stack[0]), p.Stack[0])
			assert.Equal(t, 0, len(p.Errors))
			if !assert.Equal(t, 1, len(p.Stack)) {
				spew.Dump(p.Stack)
			}
			//
			_, err := Parse(Build(p.Stack[0]))
			assert.NoError(t, err)
		}
	}
}

func TestParse(t *testing.T) {
	// p := &MiniQueryPeg{Tree: &Tree{}, Pretty: true, Buffer: `profile.age > 10`}
	p := &MiniQueryPeg{Tree: &Tree{}, Pretty: true, Buffer: "a > 1 and  (  b != 2 or name like '%wener%') or ( age == 0 )"}
	assert.NoError(t, p.Init())
	assert.NoError(t, p.Parse())
	fmt.Println("PrintSyntaxTree")
	// p.PrintSyntaxTree()
	p.Execute()

	for _, v := range p.Stack {
		fmt.Println(v)
	}
}

package entmq_test

import (
	"testing"

	"github.com/wenerme/go-miniquery/entmq"

	"github.com/stretchr/testify/assert"
)

// need to generate schema
// func TestEntQLDemo(t *testing.T) {
// 	graph := ent.GetSQLSchemaGraph()
// 	n := ent.GetSQLSchemaNodeByType("Account")
//
// 	{
// 		base := sql.Dialect(dialect.Postgres).
// 			Select().
// 			From(sql.Table(n.Table))
// 		err := graph.EvalP("Account", entql.HasEdge("owningUser"), base)
// 		if err != nil {
// 			panic(err)
// 		}
// 		s, args := base.Clone().Query()
// 		fmt.Println("query\n", s, "\n", "args\n", args)
// 	}
// 	{
// 		base := sql.Dialect(dialect.Postgres).
// 			Select().
// 			From(sql.Table(n.Table))
// 		err := graph.EvalP("Account", entql.HasEdgeWith("owningUser", entql.EQ(entql.F("username"), &entql.Value{V: "wener"})), base)
// 		if err != nil {
// 			panic(err)
// 		}
// 		s, args := base.Clone().Query()
// 		fmt.Println("query\n", s, "\n", "args\n", args)
// 	}
// 	{
// 		base := sql.Dialect(dialect.Postgres).
// 			Select().
// 			From(sql.Table(n.Table))
// 		err := graph.EvalP("Account", entql.HasEdgeWith("owningUser", entql.HasEdgeWith("department", entql.EQ(entql.F("name"), &entql.Value{V: "Test"}))), base)
// 		if err != nil {
// 			panic(err)
// 		}
// 		s, args := base.Clone().Query()
// 		fmt.Println("query\n", s, "\n", "args\n", args)
// 	}
// }

func TestQLBuilder(t *testing.T) {
	for _, test := range []struct {
		Q string
		E string
	}{
		{Q: "a>10", E: "a > 10"},
		{Q: "a between 1 and 3", E: "a >= 1 && a <= 3"},
		{Q: "a not between 1 and 3", E: "a < 1 && a > 3"},
		{Q: "true and false", E: "true && false"},
		{Q: "a in [1 , 2 , 3]", E: "a in [1,2,3]"},
	} {
		b := &entmq.MiniQLToEntQLBuilder{
			Query: test.Q,
		}
		ql, err := b.Build()
		assert.NoError(t, err)
		assert.Equal(t, test.E, ql.String())
	}
}

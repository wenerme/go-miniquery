# go-miniquery

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report Card][report-card-img]][report-card]

[doc-img]: https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square
[doc]: https://pkg.go.dev/github.com/wenerme/go-miniquery?tab=doc
[ci-img]: https://github.com/wenerme/go-miniquery/actions/workflows/ci.yml/badge.svg
[ci]: https://github.com/wenerme/go-miniquery/actions/workflows/ci.yml
[cov-img]: https://codecov.io/gh/wenerme/go-miniquery/branch/main/graph/badge.svg
[cov]: https://codecov.io/gh/wenerme/go-miniquery/branch/main
[report-card-img]: https://goreportcard.com/badge/github.com/wenerme/go-miniquery
[report-card]: https://goreportcard.com/report/github.com/wenerme/go-miniquery

SQL Where like __safe__ filter expression for entql and gorm

## gorm

- use reflect to get model graph
- support join relation

```go
package main

import "gorm.io/gorm"

func TestMiniQuery() {
	var db *gorm.DB
	// build to username = ? and Profile__age > ?
	db.Model(User{}).Scopes(ApplyMiniQuery(`username="wener" && Profile.age > 18`)).Rows()
}
```

## entql

- generate entsql - recommend
    - depends on generated graph - will validate when build
    - support edge filter
- generate entql
    - function is not as complete as entsql
    - do not depends on generated graph - validate when entql to sql

```go
package main

import "entgo.io/ent/dialect/sql"

func TestEntSQL() {
	var s *sql.Selector
	s.Where(sql.P(func(builder *sql.Builder) {
		b := &entmq.MiniQLToEntSQLBuilder{
			Node:        node,          // *sqlgraph.Node
			Graph:       schemaGraph,   // *sqlgraph.Schema
			QueryString: query,
		}
		builder.Join(b)
		err := b.Err()
		if err != nil {
			panic(errors.Wrap(err, "invalid query"))
		}
	}))
}
```

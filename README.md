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

SQL Where like filter express for entql and gorm


```go
package main

import 	"gorm.io/gorm"

func TestMiniQuery(){
	var db *gorm.DB
	// build to username = ?
	db.Model(User{}).Scopes(ApplyMiniQuery(`username="wener"`)).Rows()
}
```

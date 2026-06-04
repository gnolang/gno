// Package benchstore provides storage benchmarks for the GnoVM store layer.
//
// Run with:
//
//	go test -bench=. ./gnovm/cmd/benchstore/ -benchmem -timeout=30m
package benchstore

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// Ensure gnolang amino package is registered.
var _ = gno.Package

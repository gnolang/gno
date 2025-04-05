//go:build !benchmarkingops && !benchmarkingstorage

package main

import "testing"

func init() {
	if !testing.Testing() {
		panic("build tags benchmarkingops or benchmarkingstorage are required for measuring benchmarks")
	}
}

//go:build !benchmarkingops && !benchmarkingstorage && !benchmarkingnative

package main

import "testing"

func init() {
	if !testing.Testing() {
		panic("build tags benchmarkingops or benchmarkingstorage or benchmarkingnative are required for measuring benchmarks")
	}
}

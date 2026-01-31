//go:build !benchmarkingops && !benchmarkingstorage && !benchmarkingnative && !benchmarkingpreprocess

package main

import "testing"

func init() {
	if !testing.Testing() {
		panic("build tags benchmarkingops or benchmarkingstorage or benchmarkingnative or benchmarkingpreprocess are required for measuring benchmarks")
	}
}

//go:build !benchmarkingops && !benchmarkingstorage

package main

func init() {
	panic("build tags benchmarkingops or benchmarkingstorage are required for measuring benchmarks")
}

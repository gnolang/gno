// https://github.com/gnolang/gno/issues/3751
package main

import "testing"

func Add(a, b int) int {
	return a + b
}

func TestAdd(t *testing.T) {
	go Add(1, 1)
}

// Error:
// parse_err2.gno:11:2: goroutines are not permitted

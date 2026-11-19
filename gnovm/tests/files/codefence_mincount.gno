// MAXALLOC: 500000000
package main

import (
	"chain/markdown"
	"strings"
)

// A minCount past the cap must not size the fence, while a backtick run longer
// than the cap still gets a fence that outruns it.
func main() {
	println(len(markdown.CodeFence("", 1<<24)))
	println(len(markdown.CodeFence(strings.Repeat("`", 2000), 1<<24)))
}

// Output:
// 1024
// 2001

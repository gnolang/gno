package main

import "strings"

func (b *strings.Builder) Hello() string { return "" }

func main() {}

// Error:
// main/method_nonlocal.gno:5:9-25: cannot define new methods on non-local type strings.Builder

// TypeCheckError:
// main/method_nonlocal.gno:5:10: cannot define new methods on non-local type strings.Builder

package main

// Tests cross-file transitive dependency through a function.
// extern/initorder_crossfn: a.gno has A = GetB(), b.gno has B = 42 and
// func GetB() int { return B }.
// findDependentNames must follow GetB (a FuncDecl in b.gno) into its body
// using the correct file context (subFn = b.gno) to find that A depends on B.
// Init order within the package: B, then A.

import "filetests/extern/initorder_crossfn"

func main() { println(initorder_crossfn.A) }

// Output:
// 42

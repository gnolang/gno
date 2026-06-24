package main

// Tests cross-file transitive dependency through a function.
// extern/initorder_crossfn: a.gno has A = GetB(), b.gno has B = 42 and
// func GetB() int { return B }.
// During preprocessing of b.gno, addDependencyToTopDecl records that GetB
// uses B, so A's ATTR_DECL_DEPS includes both GetB and (transitively) B.
// Init order within the package: B (no deps), then A (deps GetB and B).

import "filetests/extern/initorder_crossfn"

func main() { println(initorder_crossfn.A) }

// Output:
// 42

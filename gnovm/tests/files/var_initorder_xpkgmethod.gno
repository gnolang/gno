package main

// Tests that findDependentNames does not recurse into method bodies defined
// in imported packages. initorder_xpkgmethod.T.GetB references that
// package's variable B; without the same-package guard, findDependentNames
// would recurse into GetB with pn=main and then panic trying to resolve "B"
// in main's fileset. With the guard, the foreign body is skipped.
//
// Also verifies that the intra-package init order of the imported package
// itself is correct: B must be initialised before A inside
// initorder_xpkgmethod (A = T{}.GetB() = 42).

import "filetests/extern/initorder_xpkgmethod"

func main() {
	println(initorder_xpkgmethod.A)
}

// Output:
// 42

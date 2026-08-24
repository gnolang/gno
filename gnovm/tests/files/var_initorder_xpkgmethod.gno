package main

// Tests that addDependencyToTopDecl does not record dependencies into method
// bodies defined in imported packages. initorder_xpkgmethod.T.GetB references
// that package's variable B; without the same-package guard (dt.PkgPath ==
// ctxpn.PkgPath), the dep would be recorded with pn=main and resolveDeclDep
// would panic trying to find "B" in main's fileset. With the guard, the
// foreign method body is skipped.
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

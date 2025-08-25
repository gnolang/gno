// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/r/demo/tests/ext"

func main() {
	ext.GetMyInt().Modify()
	println(ext.GetRoot())
}

// Output:
// (struct{(1 int)} gno.land/r/demo/tests/ext.MyMutator)

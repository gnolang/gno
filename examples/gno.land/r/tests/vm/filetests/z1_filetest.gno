// PKGPATH: gno.land/r/tests/vm/filetests/z1_filetest

package z1_filetest

import (
	tests "gno.land/r/tests/vm"
)

func main(cur realm) {
	println(tests.Counter(cross2(cur)))
	tests.IncCounter(cross2(cur))
	println(tests.Counter(cross2(cur)))
}

// Output:
// 0
// 1

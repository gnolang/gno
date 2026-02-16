// Regression test for the bug fixed in https://github.com/gnolang/gno/pull/5019

package main

import "filetests/extern/copytest"

func main() {
	_ = append(copytest.ByteSlice[:0], []byte("XXX")...)
	println(copytest.ByteSlice)
}

// Error:
// cannot modify external-realm or non-realm object

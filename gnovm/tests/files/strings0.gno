// PKGPATH: gno.land/r/demo/main

package main

import "filetests/extern/replacer"

func main(cur realm,) {
	s := replacer.Replacer.Replace("Hello")
	println(s)
}

// Output:
// H3ll0

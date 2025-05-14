// PKGPATH: gno.land/r/demo/main

package main

import "github.com/gnolang/gno/_test/replacer"

func main() {
	crossing()

	s := replacer.Replacer.Replace("Hello")
	println(s)
}

// Output:
// H3ll0

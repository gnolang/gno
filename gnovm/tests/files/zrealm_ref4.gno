// PKGPATH: gno.land/r/test
package test

import "fmt"

// Realm filetest: ref of a nil interface{} var.

var i interface{}

func main(cur realm) {
	p := &i
	fmt.Printf("%T\n", p)
}

// Output:
// *interface {}

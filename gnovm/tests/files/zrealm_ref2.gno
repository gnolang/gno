// PKGPATH: gno.land/r/test
package test

import "fmt"

// Realm filetest: ref of an interface{} var holding a concrete value.

var i interface{}

func init() {
	i = 42
}

func main(cur realm) {
	p := &i
	fmt.Printf("%T\n", p)
}

// Output:
// *interface {}

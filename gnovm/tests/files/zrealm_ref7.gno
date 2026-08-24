// PKGPATH: gno.land/r/test
package test

import "fmt"

// Realm filetest: ref of a slice element that is an interface type.

var s []interface{}

func init() {
	s = []interface{}{1, "two", 3.0}
}

func main(cur realm) {
	p := &s[0]
	fmt.Printf("%T\n", p)
	*p = "replaced"
	fmt.Println(s[0])
}

// Output:
// *interface {}
// replaced

// PKGPATH: gno.land/r/test
package test

import "fmt"

// Realm filetest: ref of a struct field that is an interface type.

type S struct {
	F interface{}
}

var s S

func init() {
	s.F = 42
}

func main(cur realm) {
	p := &s.F
	fmt.Printf("%T\n", p)
	*p = "hello"
	fmt.Println(s.F)
}

// Output:
// *interface {}
// hello

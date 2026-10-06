// PKGPATH: gno.land/r/test
package test

import "fmt"

// Realm filetest: function takes *interface{} parameter.
// Tests that the pointer type is checked correctly
// when passed to a function expecting *interface{}.

var i interface{}

func init() {
	i = 100
}

func setViaPtr(p *interface{}, val interface{}) {
	*p = val
}

func main(cur realm) {
	setViaPtr(&i, "changed")
	fmt.Println(i)
}

// Output:
// changed

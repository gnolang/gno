// PKGPATH: gno.land/r/test
package test

// Realm filetest: ref of an interface{} var, mutation through pointer.

var i interface{}

func init() {
	i = 42
}

func main(cur realm) {
	p := &i
	*p = "hello"
	println(i)
}

// Output:
// hello

// PKGPATH: gno.land/r/test
package test

// Realm filetest: ref of a struct field via pointer in a realm.
// This works correctly since the struct type is concrete.

type MyStruct struct {
	A int
}

var s MyStruct

func init() {
	s.A = 10
}

func main(cur realm) {
	p := &s
	p.A = 20
	println(s)
}

// Output:
// (struct{(20 int)} gno.land/r/test.MyStruct)

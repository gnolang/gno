// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (*foo) Foo() { println("hello from crossrealm_test") }

var f1 = foo{name: "1"}
var f2 = foo{name: "2"}
var f3 = foo{name: "3"}

var arr []crossrealm.Fooer // unreal array

func init() {
	arr = append(arr, &f1)
	arr = append(arr, &f2)
	arr = append(arr, &f3)
}

func main() {
	crossrealm.SetSlice(arr)
	println("ok")
}

// Output:
// &(struct{("1" string)} gno.land/r/crossrealm_test.foo)
// &(struct{("2" string)} gno.land/r/crossrealm_test.foo)
// &(struct{("3" string)} gno.land/r/crossrealm_test.foo)
// ok

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/slice"]
// switchrealm["gno.land/r/crossrealm_test"]

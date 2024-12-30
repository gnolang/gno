// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type Foo struct {
}

var a = Foo{}

var f = func() bool {
	println("callback")
	return true
}

func main() {
	println("ok")
}

// Output:
// ok

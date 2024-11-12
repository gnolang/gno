// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type Bar struct {
	A int
}

var b0 Bar

func init() {
	b0 = Bar{A: 1}
}

func main() {
	print(".")
}

// Output:
// .

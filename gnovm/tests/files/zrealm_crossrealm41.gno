// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type local struct {
	name string
}

var ll local // this is attached as a zero value

// XXX, consider more about this.
// var sll []crossrealm.bar
// var pll *crossrealm.bar

func main() {
	l := local{name: "a"}
	println(l)
}

// Output:
// (struct{("a" string)} gno.land/r/crossrealm_test.local)

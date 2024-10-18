// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

var ptr *int

func init() {
	ptr = new(int)
}

func main() {
	*ptr = 2
	println(*ptr)
}

// Output:
// 2

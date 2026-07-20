// PKGPATH: gno.land/r/demo/mylib/v2/v3

package v2

func main() {
	println("should fail")
}

// Error:
// package path "gno.land/r/demo/mylib/v2/v3" must not end with consecutive version suffixes

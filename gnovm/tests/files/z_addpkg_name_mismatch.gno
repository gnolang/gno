// PKGPATH: gno.land/r/demo/goodbye

package hello

func main() {
	println("should fail")
}

// Error:
// package name "hello" does not match path element "goodbye"

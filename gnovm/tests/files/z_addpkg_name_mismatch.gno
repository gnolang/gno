// PKGPATH: gno.land/r/demo/goodbye

package hello

func main() {
	println("should fail")
}

// Error:
// expected package name [goodbye] but got [hello]

// TypeCheckError:
// gno.land/r/demo/goodbye/.gnobuiltins.gno:1:1: package goodbye; expected package hello

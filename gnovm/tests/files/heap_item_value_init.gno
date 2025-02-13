// PKGPATH: gno.land/r/test
package test

type S struct {
	A *int
}

var a, b *S

func init() {
	a = new(S) // heapItemValue new real
	a.A = new(int)
	*a.A = 4
}

func main() {
	b = a
}

// Realm:
// switchrealm["gno.land/r/test"]

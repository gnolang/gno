// PKGPATH: gno.land/r/test
package test

// Verify that deletion of a non-existing key does not mark the map as dirty.
var m map[string]int

func init() {
	m = map[string]int{"a": 1, "b": 2, "c": 3}
}

func main(cur realm) {
	delete(m, "missing") // no-op; should not dirty the map
	println("len:", len(m))
}

// Output:
// len: 3

// Realm:
// finalizerealm["gno.land/r/test"]

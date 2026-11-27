// PKGPATH: gno.land/r/test
package test

// A crossing function copies a stored array of interfaces without reading
// its element, clears the source and returns the copy. The finalize at the
// crossing return deletes the element the source no longer references, so
// the copy must not share it: it would come back as a ref to a deleted object.

type S struct{ F int }

var arr [1]any

func init() {
	arr[0] = S{1}
}

func Take(cur realm) [1]any {
	b := arr // element not read: still a RefValue
	arr[0] = nil
	return b
}

func main(cur realm) {
	b := Take(cross(cur))
	println(b[0].(S).F)
}

// Output:
// 1

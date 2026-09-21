// PKGPATH: gno.land/r/curpersiststruct
package curpersiststruct

// A struct field of realm type stashed into a top-level var must trip the
// persistence check via the realm's backing *HeapItemValue.
type Holder struct {
	R realm
}

var H *Holder

func main(cur realm) {
	H = &Holder{R: cur}
	println("assignment ok, awaiting finalize panic")
}

// Error:
// cannot persist realm value: realm values are ephemeral and tied to a call frame

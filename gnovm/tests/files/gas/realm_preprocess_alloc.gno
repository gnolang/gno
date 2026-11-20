// PKGPATH: gno.land/r/test/realm_preprocess_alloc
package realm_preprocess_alloc

// make allocates while the realm package is preprocessed.
var buf = make([]byte, 10)

func main(cur realm) {
	_ = string(buf)
}

// Gas:
// 15947

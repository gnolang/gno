// PKGPATH: gno.land/r/map_nan_key
package map_nan_key

import "math"

// Composite map key: a NaN float64 field BEFORE an object-bearing struct
// field. ComputeMapKey's struct loop returns early on the first NaN (the
// isNaN abort), so it never reaches fillValueTV on the later field. On
// master, the eager vmap build relied on ComputeMapKey to deep-resolve every
// stored key's RefValue children as a side effect, so this later field was
// NEVER visited and stayed an unresolved RefValue after a store round-trip:
// master printed it as `ref(...)`. fillMapKeyRefs has no NaN abort — it
// walks the whole key regardless — so this branch prints the expanded
// struct instead. That expanded form is the CORRECT value (same bug class
// gnovm/tests/files/map39b.gno guards), but it is an observable output
// change for this narrow shape, not merely an unmetered load-time detail.

type inner struct {
	n int
}

type key struct {
	f  float64
	in inner
}

var (
	k key
	m map[key]int
)

func init() {
	k = key{f: math.NaN(), in: inner{n: 7}}
	m = map[key]int{}
	m[k] = 1
}

func main(cur realm) {
	for sk, sv := range m {
		println("key:", sk, "value:", sv)
	}
}

// Output:
// key: (struct{(NaN float64),(struct{(7 int)} gno.land/r/map_nan_key.inner)} gno.land/r/map_nan_key.key) value: 1

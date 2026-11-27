// Package recoverfv panics from a library method and runs callbacks, for the
// recover_fv file tests.
package recoverfv

type C struct {
	m map[string]any
	n int
}

func New() *C { return &C{m: map[string]any{"a": 1}} }

// Set panics on the type assertion inside the map range.
func (c *C) Set(v string) {
	for _, x := range c.m {
		_ = x.(string)
	}
}

// Each calls f once per element.
func Each(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

var shared = New()

// Shared returns an object the library itself stores.
func Shared() *C { return shared }

// Run calls f; on the shared object it runs under the library's realm.
func (c *C) Run(f func()) { f() }

// RunAny is Run with a result.
func (c *C) RunAny(f func() any) any { return f() }

// Mk returns a library closure that recovers a panic of the shared object.
func Mk() func() any {
	return func() (rec any) {
		defer func() { rec = recover() }()
		shared.Set("x")
		return
	}
}

// Mk2 nests one more library closure between the recover and the panic.
func Mk2() func() any {
	inner := func() { shared.Set("x") }
	mid := []func(){func() { inner() }}
	return func() (rec any) {
		defer func() { rec = recover() }()
		mid[0]()
		return
	}
}

// DeferPanic panics from a deferred function.
func (c *C) DeferPanic() {
	defer func() { panic("deferred") }()
}

// Make returns an object made while the method runs.
func (c *C) Make() *C { return New() }

// RunRec runs f and recovers inside the library method.
func (c *C) RunRec(f func()) (rec any) {
	defer func() { rec = recover() }()
	f()
	return
}

// Nop does nothing.
func (c *C) Nop() {}

// Deep returns a closure that recurses d frames, then calls the shared
// object's Nop n times.
func Deep(d, n int) func() {
	var rec func(int)
	rec = func(k int) {
		if k > 0 {
			rec(k - 1)
			return
		}
		for i := 0; i < n; i++ {
			shared.Nop()
		}
	}
	return func() { rec(d) }
}

// H is a helper whose method recovers a panic of the shared object.
type H struct{ k int }

func (h *H) Try() (rec any) {
	defer func() { rec = recover() }()
	shared.Set("x")
	return
}

// StoredMk is a closure the library itself stores.
var StoredMk = func() (rec any) {
	defer func() { rec = recover() }()
	shared.Set("x")
	return
}

// StoredNop calls the shared object twice: once cleanly, then with a panic.
var StoredNop = func() (rec any) {
	defer func() { rec = recover() }()
	shared.Nop()
	shared.Set("x")
	return
}

// StoredNested panics two closures deep.
var StoredNested = func() (rec any) {
	defer func() { rec = recover() }()
	inner := func() { shared.Set("x") }
	mid := func() { inner() }
	mid()
	return
}

// N reads the counter Inc writes.
func (c *C) N() int { return c.n }

// Inc writes the object's counter.
func (c *C) Inc() any { c.n++; return nil }

// StoredBoom panics on the shared object without recovering.
var StoredBoom = func() { shared.Set("x") }

// StoredInc writes the shared object.
var StoredInc = func() { shared.n++ }

// StoredRun calls f from a closure the library stores.
var StoredRun = func(f func() any) any { return f() }

// StoredTry is StoredMk under another name, for routes that keep it.
var StoredTry = func() (rec any) {
	defer func() { rec = recover() }()
	shared.Set("x")
	return
}

// MkIn returns a closure made inside a method of c that recovers c's panic.
func (c *C) MkIn() func() any {
	return func() (rec any) {
		defer func() { rec = recover() }()
		c.Set("x")
		return
	}
}

// Inc writes h.
func (h *H) Inc() any { h.k++; return nil }

// K reads h.
func (h *H) K() int { return h.k }

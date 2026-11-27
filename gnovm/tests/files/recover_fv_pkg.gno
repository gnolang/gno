// PKGPATH: gno.land/p/demo/tests/recoverfvpkg
package recoverfvpkg

type C struct{ m map[string]any }

func New() *C { return &C{m: map[string]any{"a": 1}} }

func (c *C) Set(v string) {
	for _, x := range c.m {
		_ = x.(string)
	}
}

// A panic raised through a function value is recovered by the caller's defer,
// as a direct call is.
func main() {
	c := New()
	println("direct:", try(func() { c.Set("x") }) != nil)
	ops := []func(){func() { c.Set("x") }}
	println("via value:", try(ops[0]) != nil)
	println("via value in literal:", func() (rec any) {
		defer func() { rec = recover() }()
		ops[0]()
		return
	}() != nil)
	println("OUTCOME caught")
}

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

// Output:
// direct: true
// via value: true
// via value in literal: true
// OUTCOME caught

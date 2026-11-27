// PKGPATH: gno.land/p/demo/tests/recoverfvpkgnest
package recoverfvpkgnest

type C struct{ m map[string]any }

func New() *C { return &C{m: map[string]any{"a": 1}} }

func (c *C) Set(v string) {
	for _, x := range c.m {
		_ = x.(string)
	}
}

func (c *C) DeferPanic() {
	defer func() { panic("deferred") }()
}

func (c *C) Re() {
	defer func() {
		r := recover()
		panic("re:" + r.(error).Error())
	}()
	c.Set("x")
}

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

// Panics through nested stored closures, a method value, a deferred call
// and a re-panic are recovered in a pure package, as on master.
func main() {
	c := New()
	l3 := []func(){func() { c.Set("x") }}
	l2 := []func(){func() { l3[0]() }}
	l1 := []func(){func() { l2[0]() }}
	println("nested:", try(l1[0]))
	mid := []func() any{func() any { return try(l3[0]) }}
	println("mid:", mid[0]())
	mv := []func(string){c.Set}
	println("method value:", try(func() { mv[0]("x") }))
	dp := []func(){func() { c.DeferPanic() }}
	println("deferred:", try(dp[0]))
	re := []func(){func() { c.Re() }}
	println("re:", try(re[0]))
	dd := []func(){func() { defer c.Set("x") }}
	println("defer call:", try(dd[0]))
	println("end")
	println("OUTCOME caught")
}

// Output:
// nested: int is not of type string
// mid: int is not of type string
// method value: int is not of type string
// deferred: deferred
// re: re:int is not of type string
// defer call: int is not of type string
// end
// OUTCOME caught

// PKGPATH: gno.land/r/zrealm_fvshare_sweep
// A writer put in saved state by each construct, then called
// back from that slot in the same crossing call, must be refused.
// The control is a literal lent in place to this realm's own try.
package zrealm_fvshare_sweep

type S struct{ f func() }
type F func()

func (f F) Call() { f() }

var (
	count int
	G     func()
	H     func()
	L     []func()
	M     map[string]func()
	PS    *S
	VS    S
	A     [1]func()
	B     F
	BM    func()
)

func try(name string, f func()) {
	defer func() {
		r := recover()
		println(name, "refused:", r != nil, "count:", count)
	}()
	f()
}

func named(w func()) (r func()) {
	defer func() { G = r }()
	r = w
	return
}

func main(cur realm) {
	mk := func() func() { return func() { count++ } }
	L = make([]func(), 0, 4)
	M = map[string]func(){}
	PS = &S{}

	try("control local", func() { count++ })
	w := mk()
	G = w
	try("assign", func() { G() })
	w, H = H, mk()
	try("swap", func() { H() })
	locals := []func(){mk()}
	for _, G = range locals {
	}
	try("range assign", func() { G() })
	L = append(L, mk())
	try("append in place", func() { L[0]() })
	more := []func(){mk(), mk()}
	L = append(L, more...)
	try("append spread in place", func() { L[1]() })
	L = append(L, mk(), mk(), mk(), mk())
	try("append grow", func() { L[5]() })
	copy(L, []func(){mk()})
	try("copy", func() { L[0]() })
	M["k"] = mk()
	try("map assign", func() { M["k"]() })
	M = map[string]func(){"lit": mk()}
	try("map literal", func() { M["lit"]() })
	PS.f = mk()
	try("ptr struct field", func() { PS.f() })
	VS = S{f: mk()}
	try("struct value", func() { VS.f() })
	A = [1]func(){mk()}
	try("array value", func() { A[0]() })
	B = F(mk())
	try("named func type", func() { B() })
	BM = F(mk()).Call
	try("bound method", func() { BM() })
	named(mk())
	try("named result in defer", func() { G() })
	var i any = mk()
	switch x := i.(type) {
	case func():
		G = x
	}
	try("type switch", func() { G() })
}

// Output:
// control local refused: false count: 1
// assign refused: true count: 1
// swap refused: true count: 1
// range assign refused: true count: 1
// append in place refused: true count: 1
// append spread in place refused: true count: 1
// append grow refused: true count: 1
// copy refused: true count: 1
// map assign refused: true count: 1
// map literal refused: true count: 1
// ptr struct field refused: true count: 1
// struct value refused: true count: 1
// array value refused: true count: 1
// named func type refused: true count: 1
// bound method refused: true count: 1
// named result in defer refused: true count: 1
// type switch refused: true count: 1

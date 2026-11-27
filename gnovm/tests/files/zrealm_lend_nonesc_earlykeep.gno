// PKGPATH: gno.land/r/tests/vm/lendnonescearlykeep
package lendnonescearlykeep

// Each kN keeps f through a literal that preprocess handles early (converted
// to a named func type, or in a range, switch, case, shift or append), then
// calls the kept copy while its frame is live: f must not be lent there.

type Fr func() func()

var count int

func pick(h Fr) func()                      { return h() }
func pickU(h func() func()) func()          { return h() }
func pick2(p *func(), h Fr) int             { *p = h(); return 1 }
func pick2U(p *func(), h func() func()) int { *p = h(); return 1 }

func k1(f func()) { g := pick(Fr(func() func() { return f })); g() }

func k2(f func()) {
	g := pick(Fr(func() func() {
		h := f
		f := 0
		_ = f
		return h
	}))
	g()
}

func k3(f func()) {
	for _, g := range []func(){pick(Fr(func() func() { return f }))} {
		g()
	}
}

func k4(f func()) {
	var g func()
	switch pick2(&g, Fr(func() func() { return f })) {
	}
	g()
}

func k5(f func()) {
	var g func()
	switch 1 {
	case pick2(&g, Fr(func() func() { return f })):
	}
	g()
}

func k6(f func()) {
	var g func()
	x := 1 << pick2(&g, Fr(func() func() { return f }))
	_ = x
	g()
}

func k7(f func()) {
	var gs []func()
	gs = append(gs, pick(Fr(func() func() { return f })))
	gs[0]()
}

func k8(f func()) {
	for _, g := range []func(){pickU(func() func() { return f })} {
		g()
	}
}

func k9(f func()) {
	var g func()
	switch pick2(&g, Fr(func() func() {
		h := f
		f := 0
		_ = f
		return h
	})) {
	}
	g()
}

func k10(f func()) {
	var g func()
	var n uint = 1
	x := pick2U(&g, func() func() { return f }) << n
	_ = x
	g()
}

func k11(f func()) {
	var gs []func()
	gs = append(gs, pickU(func() func() { return f }))
	gs[0]()
}

func main(cur realm) {
	for i := 0; i < 11; i++ {
		r := try(i)
		println("k", i+1, count, r != nil)
	}
}

func try(i int) (r any) {
	defer func() { r = recover() }()
	switch i {
	case 0:
		k1(func() { count++ })
	case 1:
		k2(func() { count++ })
	case 2:
		k3(func() { count++ })
	case 3:
		k4(func() { count++ })
	case 4:
		k5(func() { count++ })
	case 5:
		k6(func() { count++ })
	case 6:
		k7(func() { count++ })
	case 7:
		k8(func() { count++ })
	case 8:
		k9(func() { count++ })
	case 9:
		k10(func() { count++ })
	case 10:
		k11(func() { count++ })
	}
	return nil
}

// Output:
// k 1 0 true
// k 2 0 true
// k 3 0 true
// k 4 0 true
// k 5 0 true
// k 6 0 true
// k 7 0 true
// k 8 0 true
// k 9 0 true
// k 10 0 true
// k 11 0 true

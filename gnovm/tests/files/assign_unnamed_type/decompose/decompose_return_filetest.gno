package main

// A value that differs from its target's type only by name is converted when it
// comes from `return f()`, `g(f())`, a comma-ok assignment or a range assignment.

type nmap map[string]int

func (nmap) M() string { return "nmap" }

type nfunc func() int

type T struct{ A int }

func (t T) Get() int { return t.A }

func unamed() (map[string]int, map[string]int) { return map[string]int{"a": 1}, nil }
func named() (nmap, nmap)                      { return nmap{"a": 1}, nil }
func structs() (struct{ A int }, int)          { return struct{ A int }{7}, 8 }
func funcs() (func() int, int)                 { return func() int { return 9 }, 0 }

func plain() (nmap, nmap)                         { return unamed() }
func toUnnamed() (map[string]int, map[string]int) { return named() }
func toT() (T, int)                               { return structs() }
func toFunc() (nfunc, int)                        { return funcs() }
func toAny() (any, any)                           { return named() }

func namedRes() (a, b nmap) { return unamed() }

func captured() (a, b nmap) {
	defer func() { println("captured", kind(a), a.M()) }()
	return unamed()
}

func deferred() (nmap, nmap) {
	defer func() {}()
	return unamed()
}

type R struct{}

func (R) method() (nmap, nmap) { return unamed() }

func (R) bound(a, _ nmap) { println("bound", kind(a), a.M()) }

func param(a, b nmap) { println("param", kind(a), a.M(), kind(b)) }

func vararg(xs ...nmap) { println("vararg", kind(xs[0]), kind(xs[1])) }

func kind(x any) string {
	switch x.(type) {
	case nmap:
		return "nmap"
	case map[string]int:
		return "map"
	case T:
		return "T"
	case nfunc:
		return "nfunc"
	}
	return "other"
}

func main() {
	a, b := plain()
	println("plain", kind(a), kind(b), a.M(), b == nil)

	c, d := toUnnamed()
	println("toUnnamed", kind(c), kind(d), d == nil)

	t, n := toT()
	println("toT", kind(t), t.Get(), n)

	f, _ := toFunc()
	println("toFunc", kind(f), f())

	e, g := toAny()
	println("toAny", kind(e), kind(g))

	h, _ := namedRes()
	println("namedRes", kind(h), h.M())

	i, _ := captured()
	println("captured", kind(i))

	j, _ := deferred()
	println("deferred", kind(j), j.M())

	k, _ := R{}.method()
	println("method", kind(k), k.M())

	l, _ := func() (nmap, nmap) { return unamed() }()
	println("closure", kind(l), l.M())

	// A comma-ok assignment.
	src := map[string]map[string]int{"k": {"a": 1}}
	var o nmap
	o, _ = src["k"]
	println("comma", kind(o), o.M())
	o, _ = src["none"]
	println("missing", kind(o), o == nil)
	var boxed any = map[string]int{"a": 1}
	o, _ = boxed.(map[string]int)
	println("assert", kind(o), o.M())

	// A result spread over parameters, a variadic one and a deferred call.
	param(unamed())
	vararg(unamed())
	R{}.bound(unamed())
	func() {
		defer param(unamed())
	}()

	// A range assignment over a slice, an array pointer and a map.
	var rv nmap
	for _, rv = range []map[string]int{{"a": 1}} {
		println("range", kind(rv), rv.M())
	}
	for _, rv = range &[1]map[string]int{{"a": 1}} {
		println("rangeptr", kind(rv), rv.M())
	}
	var rk T
	for rk, rv = range map[struct{ A int }]map[string]int{{A: 2}: {"a": 1}} {
		println("rangemap", kind(rk), rk.Get(), kind(rv))
	}
}

// Output:
// plain nmap nmap nmap true
// toUnnamed map map true
// toT T 7 8
// toFunc nfunc 9
// toAny nmap nmap
// namedRes nmap nmap
// captured nmap nmap
// captured nmap
// deferred nmap nmap
// method nmap nmap
// closure nmap nmap
// comma nmap nmap
// missing nmap true
// assert nmap nmap
// param nmap nmap nmap
// vararg nmap nmap
// bound nmap nmap
// param nmap nmap nmap
// range nmap nmap
// rangeptr nmap nmap
// rangemap T 2 nmap

package lendpickvictim

import lib "gno.land/p/demo/tests/lendpicklib"

type Vault struct{ N int }

func (v *Vault) Bump() { v.N += 1000 }

var vault = &Vault{}
var box = &lib.Box{}

func Get() *Vault { return vault }
func Total() int  { return vault.N + box.N }

// RunParam calls a caller's callback; the later := reuses the parameter.
func RunParam(cur realm, g func()) {
	g()
	g, ok := func() {}, true
	_ = ok
}

// RunParamBox: same, over a stored box.
func RunParamBox(cur realm, g func(*lib.Box)) {
	g(box)
	g, ok := func(*lib.Box) {}, true
	_ = ok
}

// RunResult: the named result is reused by :=, then filled by return.
func RunResult(cur realm, h func()) {
	pick(h)
}

func pick(h func()) (g func()) {
	g, ok := func() {}, true
	_ = ok
	defer func() { g() }()
	return lib.Pick(h)
}

// RunPlain: control, the parameter alone.
func RunPlain(cur realm, g func()) { g() }

// RunReg hands a caller's callback to a library-owned object's method.
func RunReg(cur realm, g func(*lib.Box)) { lib.Reg.Each(box, g) }
func RunRegPlain(cur realm, g func())    { lib.Reg.Run(g) }

type Fn func()

func (f Fn) Run() { f() }

type Runner interface{ Run() }

// RunIface calls a caller's value through an interface.
func RunIface(cur realm, r Runner) { r.Run() }

type Hooks struct{ F func() }

// RunField calls a caller's value through a struct field.
func RunField(cur realm, g func()) {
	h := Hooks{F: g}
	h.F()
}

// Plain is a non-crossing entry that calls its parameter.
func Plain(g func()) { g() }

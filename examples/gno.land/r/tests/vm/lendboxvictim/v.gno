package lendboxvictim

import lib "gno.land/p/demo/tests/lendboxlib"

type Vault struct{ N int }

func (v *Vault) Bump() { v.N += 1000 }

var vault = &Vault{}

func Get() *Vault { return vault }

var boxes []*lib.Box
var counts []*int
var hook func()

func Total() int {
	s := vault.N
	for _, b := range boxes {
		s += b.N
	}
	for _, c := range counts {
		s += *c
	}
	return s
}

// AddBox runs a caller-picked init over a fresh box through a library, then keeps the box.
func AddBox(cur realm, f func(*lib.Box)) {
	b := &lib.Box{}
	lib.Each([]*lib.Box{b}, f)
	boxes = append(boxes, b)
}

// AddCount: same with an escaping local int.
func AddCount(cur realm, f func(*int)) {
	n := 0
	lib.EachInt(&n, f)
	counts = append(counts, &n)
}

func SetHook(cur realm, f func()) { hook = f }

// RunHook calls the package-level hook; an unrelated local helper shares its name.
func RunHook(cur realm) {
	if false {
		hook := func() {}
		hook()
	}
	hook()
}

func RunHookPlain(cur realm) { hook() }

// RunHookLib hands the stored hook to a library.
func RunHookLib(cur realm) { lib.Run(hook) }

// RunLit calls a caller value inside a literal called where written.
func RunLit(cur realm, f func()) {
	func() { f() }()
}

// RunDeferLit: same, deferred.
func RunDeferLit(cur realm, f func()) {
	defer func() { f() }()
}

var defInit = lib.Bump
var defIntInit = lib.BumpInt

// DefInit exposes a stored library handle (made under this realm).
func DefInit() func(*lib.Box) { return defInit }
func DefIntInit() func(*int)   { return defIntInit }

var adder = lib.MakeAdder(1000)
var intAdder = lib.MakeIntAdder(1000)

// Adder exposes a stored library closure made under this realm.
func Adder() func(*lib.Box) { return adder }
func IntAdder() func(*int)  { return intAdder }

// AddBoxD: the realm's own library closure (known-OK class; checks keep fires).
func AddBoxD(cur realm, d int) {
	b := &lib.Box{}
	lib.Each([]*lib.Box{b}, lib.MakeAdder(d))
	boxes = append(boxes, b)
}

// AddBoxGuarded wraps the caller's init in a library guard before running it.
func AddBoxGuarded(cur realm, f func(*lib.Box)) {
	b := &lib.Box{}
	lib.Each([]*lib.Box{b}, lib.Guard(f))
	boxes = append(boxes, b)
}

func AddCountGuarded(cur realm, f func(*int)) {
	n := 0
	lib.EachInt(&n, lib.GuardInt(f))
	counts = append(counts, &n)
}

// UpdateGuarded runs a guarded caller callback over boxes already saved.
func UpdateGuarded(cur realm, f func(*lib.Box)) { lib.Each(boxes, lib.Guard(f)) }

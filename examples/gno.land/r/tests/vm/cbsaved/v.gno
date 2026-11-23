// Package cbsaved is a test fixture for the SAVED function-value probe: each
// exported gate makes a saved (or same-transaction) value write.
package cbsaved

import "gno.land/p/demo/tests/libtable/v0"

var Owner = "victim"

func Set(s string) { Owner = s }

type T struct{ Name string }

func (t *T) SetName(s string) { t.Name = s }

type S struct{ F func(string) }

var (
	Box    = &T{}
	Table  = []func(string){Set}
	St     = &S{F: Set}
	LB     = &libtable.Box{}
	sameTx func(string)
	Reg    []func(string)
)

// CaptureFromTable copies this realm's saved value into a fresh closure.
func CaptureFromTable(cur realm) {
	f := Table[0]
	g := func() { f("captured") }
	g()
}

// SaveThenRun stores a fresh closure in a package variable and runs it in the
// same transaction, before any finalize could make it real.
func SaveThenRun(cur realm) {
	sameTx = func(s string) { Owner = s }
	sameTx("same-tx")
}

// RegisterThenRun stores a fresh closure in a package slice and runs it back.
func RegisterThenRun(cur realm) {
	Reg = append(Reg, func(s string) { Owner = s })
	for _, f := range Reg {
		f("registered")
	}
}

// BindThenRun stores a method value in a package struct and runs it here.
func BindThenRun(cur realm) {
	St.F = Box.SetName
	St.F("bound")
}

// DeferSaved defers a call of this realm's stored value.
func DeferSaved(cur realm) {
	defer Table[0]("deferred")
}

// PanicWithSaved panics with a stored value and runs it after recovery.
func PanicWithSaved(cur realm) {
	defer func() {
		if f, ok := recover().(func(string)); ok {
			f("recovered")
		}
	}()
	panic(Table[0])
}

// GetFV returns this realm's stored value from non-crossing code.
func GetFV() func(string) { return Table[0] }

// CallThroughGetter calls the stored value a getter returned.
func CallThroughGetter(cur realm) {
	f := GetFV()
	f("getter")
}

// LibInPlace calls a library table entry in place.
func LibInPlace(cur realm) { libtable.Handlers["inc"](LB) }

// LibCopy calls a local copy of the same library table entry.
func LibCopy(cur realm) {
	f := libtable.Handlers["inc"]
	f(LB)
}

// --- R2a attack helpers ---

var local []func(string)

// MakeValue returns a fresh writer made in this crossing call.
func MakeValue(cur realm) func(string) { return func(s string) { Owner = s } }

// CallValue calls a value handed back in a later crossing call.
func CallValue(cur realm, f func(string)) { f("returned") }

// Do builds a local table and hands it to a caller's crossing callback.
func Do(cur realm, cb func(cur realm, hs []func(string))) {
	hs := []func(string){func(s string) { Owner = s }}
	cb(cross(cur), hs)
}

// RunAll calls each entry of a handed-in table.
func RunAll(cur realm, hs []func(string)) {
	for _, f := range hs {
		f("runall")
	}
}

// MakeLocal leaves a fresh table in this realm's local slot.
func MakeLocal(cur realm) {
	local = []func(string){func(s string) { Owner = s }}
}

// AttachLocal attaches the local table by reference for later.
func AttachLocal(cur realm) { Reg = local }

// RunReg calls the attached table.
func RunReg(cur realm) {
	for _, f := range Reg {
		f("reg")
	}
}

// BindFresh returns a method value of a fresh receiver.
func BindFresh(cur realm, t *T) func(string) { return t.SetName }

// CallBound calls a method value handed back earlier.
func CallBound(cur realm, f func(string)) { f("bound2") }

// CallVarg calls each variadic-packed value.
func CallVarg(cur realm, fs ...func(string)) {
	for _, f := range fs {
		f("varg")
	}
}

// AliasTable calls a stored value through a name that hides the callee.
func AliasTable(cur realm) {
	SetOwner := Table[0]
	SetOwner("alias")
}

// CallHelper passes a fresh closure to its own realm crossing helper.
func CallHelper(cur realm) {
	f := func(s string) { Owner = s }
	helper(cross(cur), f)
}

func helper(cur realm, f func(string)) { f("helper") }

package pfuncvictim

import (
	"sort"

	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/p/demo/tests/pfuncbind"
)

type Fn = func(*launderpkg.Immutable)

type Holder struct{ F Fn }

var gObj = &launderpkg.Immutable{Field: "victim-orig"}

func Read() string    { return gObj.Field }
func Reset(cur realm) { gObj.Field = "victim-orig" }

// Each runs a caller callback on the victim's own object; a /p/ callback
// handed in must run view-only.
func ApplyOne(cur realm, f Fn)            { f(gObj) }
func ApplyVar(cur realm, fs ...Fn)        { fs[0](gObj) }
func ApplySlice(cur realm, fs []Fn)       { fs[0](gObj) }
func ApplyStruct(cur realm, h Holder)     { h.F(gObj) }
func ApplyPtrStruct(cur realm, h *Holder) { h.F(gObj) }
func ApplyMap(cur realm, m map[string]Fn) { m["a"](gObj) }
func ApplyArr(cur realm, a [1]Fn)         { a[0](gObj) }
func ApplyAny(cur realm, x any)           { x.(Fn)(gObj) }
func ApplyAnySlice(cur realm, x any)      { x.([]Fn)[0](gObj) }

// Echo returns what it was given.
func EchoSlice(fs []Fn) Fn { return fs[0] }

var gStored Holder

func Store(cur realm, h Holder) { gStored = h }
func RunStored(cur realm)       { gStored.F(gObj) }

func ApplyAnon(cur realm, h struct{ F Fn }) { h.F(gObj) }

// Call runs a caller callback (a common "notify" shape).
func Call(cur realm, f func()) { f() }

// Hook is a package var: calling it reads a value, so it runs view-only.
var Hook Fn

func SetHook(cur realm, f Fn) { Hook = f }
func RunHook(cur realm)       { Hook(gObj) }

// ApplyNamed names the /p/ function in source, so it writes as the victim.
func ApplyNamed(cur realm) { launderattack.EvilWrite(gObj) }

// Hof hands a caller callback to a /p/ helper that calls it.
func Hof(cur realm, f Fn) { pfuncbind.Apply(f, gObj) }

// SortOwn sorts the victim's own slice through a stdlib function value.
func SortOwn(cur realm) { f := sort.Ints; f(gInts) }

var gInts = []int{3, 1, 2}

// Ints returns the slice SortOwn sorts.
func Ints() []int { return gInts }

// Runner is implemented by caller-chosen code; only Use's interface is named.
type Runner interface{ Run(*launderpkg.Immutable) }

var gRunner Runner

func Use(cur realm, r Runner)         { r.Run(gObj) }
func StoreRunner(cur realm, r Runner) { gRunner = r }
func UseStored(cur realm)             { gRunner.Run(gObj) }

// RunHookOn names Do, but Do then calls the caller's function value.
func RunHookOn(cur realm, h pfuncbind.Hook) { h.Do(gObj) }

// VHook is the realm's own function type; Do calls the receiver value.
type VHook func(*launderpkg.Immutable)

func (h VHook) Do(x *launderpkg.Immutable) { h(x) }

func RunVHook(cur realm, h VHook) { h.Do(gObj) }

// ApplyNC is ApplyOne without crossing: reached as a value, it runs read-only.
func ApplyNC(f Fn) { f(gObj) }

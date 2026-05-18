// Package launderrvictim is the /r/-DATA-DECLARED variant of the
// launder-game victim. Its Immutable type is declared HERE (in /r/),
// not in /p/launderpkg. This is the recommended inter-realm pattern:
// realms declare their own logic data types.
//
// The hypothesis under test: with /r/-declared logic data, the
// Attack H/I/J/K/L laundering shapes are structurally impossible.
// Tests against this victim should all fail to mutate gImm.
package launderrvictim

import "gno.land/p/demo/tests/launderpkg"

// Immutable is /r/-declared (the key difference from /r/laundervictim,
// which uses /p/launderpkg.Immutable).
type Immutable struct {
	Field string
}

// Read is /r/launderrvictim-declared, so calling it Rule-1-borrows
// m.Realm to launderrvictim.
func (i *Immutable) Read() string { return i.Field }

var gImm *Immutable

func init() {
	gImm = &Immutable{Field: "rdata-original"}
}

// GetImm hands out a pointer to gImm. Standard "victim exposes a
// pointer to its state" antipattern — but with /r/-declared data,
// the attacker should still be unable to write through it.
func GetImm() *Immutable { return gImm }

// ReadImm reads the current field for after-attack verification.
func ReadImm() string { return gImm.Field }

// UseAnyMutator boxes gImm as any and dispatches a /p/-declared
// AnyMutator. This is the dangerous shape from Attack L: victim
// boxes its own /r/-declared data through a /p/-defined interface
// that the attacker can implement.
func UseAnyMutator(m launderpkg.AnyMutator) {
	m.Run(gImm)
}

// ApplyHook dispatches a caller-supplied callback on gImm. The
// callback's parameter type is /r/launderrvictim-declared, so /p/
// packages can't supply this hook — only /r/ realms can.
func ApplyHook(h func(*Immutable)) {
	h(gImm)
}

// --- /p/-type embedded / fielded inside /r/-declared types ---
//
// The three shapes below mix /r/-declared containers with /p/-typed
// inner state. Even though the container types are /r/-declared,
// the inner /p/-typed values inherit /p/'s methods — including
// Apply-style higher-order methods that take /p/-typed callbacks.
// /p/-attacker code can supply such callbacks. The attacker reaches
// the inner /p/-value (read access works), then invokes Apply with
// a /p/-declared function pointer. Inside Apply (Rule 2 borrowed
// to /r/launderrvictim), the callback runs without any Rule 1/2
// shift (top-level /p/ fn), so the write commits under victim
// authority.

// WithEmbed embeds launderpkg.Immutable by VALUE (method promotion
// gives WithEmbed an .Apply method).
type WithEmbed struct {
	launderpkg.Immutable
}

// WithPtr has a POINTER FIELD to launderpkg.Immutable.
type WithPtr struct {
	Inner *launderpkg.Immutable
}

// WithVal has a VALUE FIELD of launderpkg.Immutable (not embedded;
// the field is named, no method promotion — but the value is still
// addressable through c.Inner).
type WithVal struct {
	Inner launderpkg.Immutable
}

var (
	gWithEmbed *WithEmbed
	gWithPtr   *WithPtr
	gWithVal   *WithVal
)

func init() {
	gWithEmbed = &WithEmbed{Immutable: launderpkg.Immutable{Field: "embed-orig"}}
	gWithPtr = &WithPtr{Inner: &launderpkg.Immutable{Field: "ptr-orig"}}
	gWithVal = &WithVal{Inner: launderpkg.Immutable{Field: "val-orig"}}
}

func GetWithEmbed() *WithEmbed { return gWithEmbed }
func GetWithPtr() *WithPtr     { return gWithPtr }
func GetWithVal() *WithVal     { return gWithVal }

func ReadEmbed() string { return gWithEmbed.Field }
func ReadPtr() string   { return gWithPtr.Inner.Field }
func ReadVal() string   { return gWithVal.Inner.Field }

// --- Methods-less /p/-type inner state ---
// launderpkg.Bare has no methods. These containers wrap Bare in
// the three field shapes. The question: does readonly taint catch
// a direct field write through the /r/-container's getter?

type WithBareEmbed struct {
	launderpkg.Bare
}

type WithBarePtr struct {
	Inner *launderpkg.Bare
}

type WithBareVal struct {
	Inner launderpkg.Bare
}

var (
	gWithBareEmbed *WithBareEmbed
	gWithBarePtr   *WithBarePtr
	gWithBareVal   *WithBareVal
)

func init() {
	gWithBareEmbed = &WithBareEmbed{Bare: launderpkg.Bare{Field: "bare-embed-orig"}}
	gWithBarePtr = &WithBarePtr{Inner: &launderpkg.Bare{Field: "bare-ptr-orig"}}
	gWithBareVal = &WithBareVal{Inner: launderpkg.Bare{Field: "bare-val-orig"}}
}

func GetWithBareEmbed() *WithBareEmbed { return gWithBareEmbed }
func GetWithBarePtr() *WithBarePtr     { return gWithBarePtr }
func GetWithBareVal() *WithBareVal     { return gWithBareVal }

func ReadBareEmbed() string { return gWithBareEmbed.Field }
func ReadBarePtr() string   { return gWithBarePtr.Inner.Field }
func ReadBareVal() string   { return gWithBareVal.Inner.Field }

// --- Composite containers holding /p/-typed elements ---
// Slices, arrays, maps of methods-less /p/-Bare values and pointers.

var (
	gBareSlice    []launderpkg.Bare
	gBarePtrSlice []*launderpkg.Bare
	gBareArr      [3]launderpkg.Bare
	gBarePtrArr   [3]*launderpkg.Bare
	gBareMap      map[string]launderpkg.Bare
	gBarePtrMap   map[string]*launderpkg.Bare
)

func init() {
	gBareSlice = []launderpkg.Bare{
		{Field: "slice0"}, {Field: "slice1"},
	}
	gBarePtrSlice = []*launderpkg.Bare{
		{Field: "ptrslice0"}, {Field: "ptrslice1"},
	}
	gBareArr = [3]launderpkg.Bare{
		{Field: "arr0"}, {Field: "arr1"}, {Field: "arr2"},
	}
	gBarePtrArr = [3]*launderpkg.Bare{
		{Field: "ptrarr0"}, {Field: "ptrarr1"}, {Field: "ptrarr2"},
	}
	gBareMap = map[string]launderpkg.Bare{
		"a": {Field: "mapA"}, "b": {Field: "mapB"},
	}
	gBarePtrMap = map[string]*launderpkg.Bare{
		"a": {Field: "ptrmapA"}, "b": {Field: "ptrmapB"},
	}
}

func GetBareSlice() []launderpkg.Bare         { return gBareSlice }
func GetBarePtrSlice() []*launderpkg.Bare     { return gBarePtrSlice }
func GetBareArr() *[3]launderpkg.Bare         { return &gBareArr }
func GetBarePtrArr() *[3]*launderpkg.Bare     { return &gBarePtrArr }
func GetBareMap() map[string]launderpkg.Bare  { return gBareMap }
func GetBarePtrMap() map[string]*launderpkg.Bare { return gBarePtrMap }

func ReadBareSlice0() string      { return gBareSlice[0].Field }
func ReadBareSlice0Then1() string { return gBareSlice[1].Field }
func ReadBarePtrSlice0() string { return gBarePtrSlice[0].Field }
func ReadBareArr0() string      { return gBareArr[0].Field }
func ReadBarePtrArr0() string   { return gBarePtrArr[0].Field }
func ReadBareMapA() string      { return gBareMap["a"].Field }
func ReadBarePtrMapA() string   { return gBarePtrMap["a"].Field }

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

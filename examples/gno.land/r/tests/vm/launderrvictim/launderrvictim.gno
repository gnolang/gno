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

// Package laundervictim is the "victim" realm in the launder-game
// tests. It exposes a package-level g that an attacker tries to
// mutate by various means. Two shapes are exposed:
//
//   - gVal is `launderpkg.Object` (value type).
//   - gPtr is `*launderpkg.Object` (pointer to a fresh Object).
//
// Both are allocated at init under this realm's context, so their
// PkgID stamp is /r/.../laundervictim. The attacker's job is to make
// the stamp not match m.Realm at the write site — by laundering the
// stamp, capturing the value, or exploiting a borrow-rule shift.
package laundervictim

import "gno.land/p/demo/tests/launderpkg"

var (
	gVal launderpkg.Object
	gPtr *launderpkg.Object
	// gImm is an Immutable: same layout as Object but no mutator
	// method in /p/launderpkg. Victim exposes a pointer to it,
	// intending "callers can read but not write."
	gImm *launderpkg.Immutable
	// gBuf is a victim-owned byte buffer (real, /r/laundervictim-stamped
	// after init). Used to probe whether a stdlib method (e.g.
	// base64.Encode) can be tricked into writing the victim's own buffer
	// when an attacker passes it as an out-parameter.
	gBuf []byte
)

func init() {
	gVal = launderpkg.Object{Field: "original-val"}
	gPtr = &launderpkg.Object{Field: "original-ptr"}
	gImm = &launderpkg.Immutable{Field: "original-imm"}
	gBuf = []byte("original-buffer!")
}

// GetVal returns g by VALUE (caller gets a copy).
func GetVal() launderpkg.Object { return gVal }

// GetPtr returns the pointer to gPtr's underlying Object. The
// returned pointer aliases the victim's persisted state.
func GetPtr() *launderpkg.Object { return gPtr }

// GetValAddr returns &gVal — a pointer to the value-typed slot.
// The returned pointer aliases the victim's persisted state.
func GetValAddr() *launderpkg.Object { return &gVal }

// GetImm returns the pointer to gImm — a *Immutable, which has no
// mutator method in /p/launderpkg. Victim's intent: callers can read
// but not write.
func GetImm() *launderpkg.Immutable { return gImm }

// GetBuf returns the victim's own byte buffer. The returned slice
// aliases the victim's persisted backing array (/r/laundervictim-stamped).
func GetBuf() []byte { return gBuf }

// ReadVal / ReadPtr / ReadImm / ReadBuf report the current values for
// after-attack verification.
func ReadVal() string { return gVal.Field }
func ReadPtr() string { return gPtr.Field }
func ReadImm() string { return gImm.Field }
func ReadBuf() string { return string(gBuf) }

// Exploiter is the interface the victim accepts. The attacker
// supplies an implementation; the victim invokes Something(g)
// passing its own g. This is the "victim hands attacker the data"
// vector.
type Exploiter interface {
	Something(launderpkg.Object)
	SomethingPtr(*launderpkg.Object)
}

// Invoke calls the attacker's methods passing the victim's g by both
// value and by pointer.
func Invoke(e Exploiter) {
	e.Something(gVal)
	e.SomethingPtr(gPtr)
}

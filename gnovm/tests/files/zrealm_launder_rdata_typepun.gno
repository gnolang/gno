// PKGPATH: gno.land/r/launderrh5
package launderrh5

// /r/-DECLARED victim — type-pun attack (Attack G variant).
//
// Attacker /r/launderrh5 declares LocalImm with the same struct
// layout as launderrvictim.Immutable plus a mutator method. Tries
// to convert *launderrvictim.Immutable to *LocalImm to invoke the
// mutator.
//
// Blocked by doOpConvert Case 1: the source value is foreign-readonly
// (resident in /r/launderrvictim's state), so converting it to any
// type other than m.Realm's own declared type panics at the
// conversion site.

import (
	"gno.land/r/tests/vm/launderrvictim"
)

type LocalImm struct {
	Field string
}

func (l *LocalImm) Set(s string) { l.Field = s }

func main(cur realm) {
	rImm := launderrvictim.GetImm()
	li := (*LocalImm)(rImm)
	li.Set("pwnd")
}

// Error:
// illegal conversion of readonly or externally stored value

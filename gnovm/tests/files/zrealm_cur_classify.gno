// PKGPATH: gno.land/r/curclassify
package curclassify

import (
	"chain/runtime/unsafe"
)

// Exercises the classification methods (IsCode / IsUser / IsUserCall /
// IsUserRun / IsEphemeral) on captured `cur realm` values, and confirms
// each agrees with the equivalent on the runtime.Realm struct surfaced
// by unsafe.CurrentRealm() / unsafe.PreviousRealm().
func main(cur realm) {
	rc := unsafe.CurrentRealm()
	rp := unsafe.PreviousRealm()
	prev := cur.Previous()

	// CurrentRealm parity — captured cur is a normal code realm.
	println("cur.IsCode == rt.IsCode:", cur.IsCode() == rc.IsCode())
	println("cur.IsUser == rt.IsUser:", cur.IsUser() == rc.IsUser())
	println("cur.IsUserCall == rt.IsUserCall:", cur.IsUserCall() == rc.IsUserCall())
	println("cur.IsUserRun == rt.IsUserRun:", cur.IsUserRun() == rc.IsUserRun())
	println("cur.IsEphemeral == rt.IsEphemeral:", cur.IsEphemeral() == rc.IsEphemeral())

	// Concrete values for cur (a code realm).
	println("cur.IsCode:", cur.IsCode())
	println("cur.IsUser:", cur.IsUser())
	println("cur.IsUserCall:", cur.IsUserCall())
	println("cur.IsUserRun:", cur.IsUserRun())
	println("cur.IsEphemeral:", cur.IsEphemeral())

	// PreviousRealm parity — chain root, EOA-shaped.
	println("prev.IsCode == rt.IsCode:", prev.IsCode() == rp.IsCode())
	println("prev.IsUser == rt.IsUser:", prev.IsUser() == rp.IsUser())
	println("prev.IsUserCall == rt.IsUserCall:", prev.IsUserCall() == rp.IsUserCall())
	println("prev.IsUserRun == rt.IsUserRun:", prev.IsUserRun() == rp.IsUserRun())

	// Concrete values for prev (EOA caller — empty pkgPath).
	println("prev.IsCode:", prev.IsCode())
	println("prev.IsUser:", prev.IsUser())
	println("prev.IsUserCall:", prev.IsUserCall())
}

// Output:
// cur.IsCode == rt.IsCode: true
// cur.IsUser == rt.IsUser: true
// cur.IsUserCall == rt.IsUserCall: true
// cur.IsUserRun == rt.IsUserRun: true
// cur.IsEphemeral == rt.IsEphemeral: true
// cur.IsCode: true
// cur.IsUser: false
// cur.IsUserCall: false
// cur.IsUserRun: false
// cur.IsEphemeral: false
// prev.IsCode == rt.IsCode: true
// prev.IsUser == rt.IsUser: true
// prev.IsUserCall == rt.IsUserCall: true
// prev.IsUserRun == rt.IsUserRun: true
// prev.IsCode: false
// prev.IsUser: true
// prev.IsUserCall: true

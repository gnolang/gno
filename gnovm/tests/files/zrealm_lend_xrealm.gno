// PKGPATH: gno.land/r/tests/vm/lendxrealm
package lendxrealm

import (
	"gno.land/r/tests/vm/lendholdother"
	"gno.land/r/tests/vm/lendholdvictim"
)

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

func main(cur realm) {
	lendholdvictim.Attacks(cross(cur))
	lendholdvictim.Report("x09 outsider runs non-crossing lend", try(func() { lendholdvictim.ViewLend() }))
	lendholdvictim.Report("x10 outsider runs kept", lendholdother.Run(cross(cur)))
	lendholdvictim.Report("x11 outsider calls kept value", try(func() { lendholdvictim.Kept()() }))
}

// Output:
// x01 other realm non-crossing 0 gno.land/r/tests/vm/lendholdvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendholdvictim) package{})).Count
// x02 other realm crossing 0 gno.land/r/tests/vm/lendholdvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendholdvictim) package{})).Count
// x03 other realm via lib 0 gno.land/r/tests/vm/lendholdvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendholdvictim) package{})).Count
// x04 other realm keeps 0 undefined
// x05 other realm runs kept 0 gno.land/r/tests/vm/lendholdvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendholdvictim) package{})).Count
// x06 bounce back crossing 0 gno.land/r/tests/vm/lendholdvictim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendholdvictim) package{})).Count
// x07 lib callback re-enters 2 undefined
// x08 view victim lends 3 undefined
// x09 outsider runs non-crossing lend 3 gno.land/r/tests/vm/lendholdvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendholdvictim) package{})).Count
// x10 outsider runs kept 3 gno.land/r/tests/vm/lendholdvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendholdvictim) package{})).Count
// x11 outsider calls kept value 3 gno.land/r/tests/vm/lendholdvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendholdvictim) package{})).Count

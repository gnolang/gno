// PKGPATH: gno.land/r/tests/vm/chaldxrealm
package chaldxrealm

import (
	"gno.land/r/tests/vm/chaldr"
	"gno.land/r/tests/vm/chaldv"
)

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

func main(cur realm) {
	chaldv.Attacks(cross(cur))
	chaldv.Report("x09 outsider runs non-crossing lend", try(func() { chaldv.ViewLend() }))
	chaldv.Report("x10 outsider runs kept", chaldr.Run(cross(cur)))
	chaldv.Report("x11 outsider calls kept value", try(func() { chaldv.Kept()() }))
}

// Output:
// x01 other realm non-crossing 0 gno.land/r/tests/vm/chaldv's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/chaldv) package{})).Count
// x02 other realm crossing 0 gno.land/r/tests/vm/chaldv's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/chaldv) package{})).Count
// x03 other realm via lib 0 gno.land/r/tests/vm/chaldv's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/chaldv) package{})).Count
// x04 other realm keeps 0 undefined
// x05 other realm runs kept 0 gno.land/r/tests/vm/chaldv's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/chaldv) package{})).Count
// x06 bounce back crossing 0 gno.land/r/tests/vm/chaldv's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldv) package{})).Count
// x07 lib callback re-enters 2 undefined
// x08 view victim lends 3 undefined
// x09 outsider runs non-crossing lend 3 gno.land/r/tests/vm/chaldv's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/chaldv) package{})).Count
// x10 outsider runs kept 3 gno.land/r/tests/vm/chaldv's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/chaldv) package{})).Count
// x11 outsider calls kept value 3 gno.land/r/tests/vm/chaldv's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/chaldv) package{})).Count

// Cross-package interface embedding. An embedded interface's EXPORTED methods
// flatten by signature, so a cross-package embed equals the explicit method
// set. An UNEXPORTED method keeps its origin-package identity through
// flattening: it stays distinct from a same-named local unexported method, a
// local type cannot satisfy another package's sealed interface, and two
// same-named unexported methods from different packages coexist as distinct
// methods. Matches Go (guards the cross-package soundness fix; see adr/pr5739).
package main

import (
	"filetests/extern/ifaceext"
	"filetests/extern/ifaceext2"
)

type Tloc struct{}

func (Tloc) sec() int { return 1 }

func eq(a, b interface{}) bool { return a == b }

func main() {
	println("exported-embed:", eq(struct{ X interface{ ifaceext.Str } }{}, struct{ X interface{ S() string } }{}))
	println("unexported-distinct:", eq(struct{ X interface{ ifaceext.Sec } }{}, struct{ X interface{ sec() int } }{}))
	var v interface{} = Tloc{}
	_, sealed := v.(interface{ ifaceext.Sec })
	println("sealed-bypass:", sealed)
	_, local := v.(interface{ sec() int })
	println("local-ok:", local)
	// two unexported sec() from different packages are distinct and coexist:
	// interface{ ifaceext.Sec; ifaceext2.Sec } has both, distinct from one.
	dup := struct {
		X interface {
			ifaceext.Sec
			ifaceext2.Sec
		}
	}{}
	println("dup-coexist==self:", eq(dup, struct {
		X interface {
			ifaceext.Sec
			ifaceext2.Sec
		}
	}{}))
	println("dup!=one:", !eq(dup, struct{ X interface{ ifaceext.Sec } }{}))
	// sec stays selectable from within its origin package, even through
	// main's re-spelled interface: the access gate keys on the method's
	// origin (ifaceext), and main->ifaceext interface conversion passes it.
	var s interface{ ifaceext.Sec } = ifaceext.SecImpl{}
	println("origin-call:", ifaceext.UseSec(s))
}

// Output:
// exported-embed: true
// unexported-distinct: false
// sealed-bypass: false
// local-ok: true
// dup-coexist==self: true
// dup!=one: true
// origin-call: 7

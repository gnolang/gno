// Cross-package interface embedding. An embedded interface's EXPORTED methods
// flatten by signature, so a cross-package embed equals the explicit method
// set. An UNEXPORTED method keeps its origin-package identity through
// flattening: it stays distinct from a same-named local unexported method,
// and a local type cannot satisfy another package's sealed interface. Matches
// Go (guards the cross-package soundness fix; see adr/pr5739).
package main

import "filetests/extern/ifaceext"

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
}

// Output:
// exported-embed: true
// unexported-distinct: false
// sealed-bypass: false
// local-ok: true

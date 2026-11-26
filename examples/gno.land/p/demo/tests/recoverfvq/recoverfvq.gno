// Package recoverfvq is a second library that recovers recoverfv's panic,
// for the recover_fv file tests.
package recoverfvq

import "gno.land/p/demo/tests/recoverfv"

// StoredQ is a closure this library stores.
var StoredQ = func() (rec any) {
	defer func() { rec = recover() }()
	recoverfv.Shared().Set("x")
	return
}

type G struct{}

// Try recovers the shared object's panic.
func (g *G) Try() (rec any) {
	defer func() { rec = recover() }()
	recoverfv.Shared().Set("x")
	return
}

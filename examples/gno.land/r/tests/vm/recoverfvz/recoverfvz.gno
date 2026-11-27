// Package recoverfvz holds a library object and plain functions that recover
// its panic, for the recover_fv file tests.
package recoverfvz

import "gno.land/p/demo/tests/recoverfv"

var obj = recoverfv.New()

// Try recovers a panic of this realm's stored library object.
func Try() (rec any) {
	defer func() { rec = recover() }()
	obj.Set("x")
	return nil
}

// Get hands Try out as a value.
func Get() func() any { return Try }

// Via runs Try inside a borrowed library method.
func Via() any { return recoverfv.Shared().RunAny(Try) }

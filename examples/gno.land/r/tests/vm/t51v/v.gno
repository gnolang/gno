package t51v

import (
	"gno.land/p/demo/tests/t51lib"
	"gno.land/r/tests/vm/t51other"
)

var count int

func Count() int { return count }

// UseRunner runs its own literal through code the caller picked.
func UseRunner(cur realm, r t51lib.Runner) { r.Run(func() { count++ }) }

// UseDirect: same interface, implementation in the victim's import tree.
func UseDirect(cur realm) {
	var r t51lib.Runner = t51lib.Direct{}
	r.Run(func() { count++ })
}

// UseOther lends its literal to another realm's crossing call.
func UseOther(cur realm) { t51other.Run(cross(cur), func() { count++ }) }

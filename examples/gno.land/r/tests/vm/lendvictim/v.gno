package lendvictim

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/r/tests/vm/lendother"
)

var count int

func Count() int { return count }

// UseRunner runs its own literal through code the caller picked.
func UseRunner(cur realm, r lendlib.Runner) { r.Run(func() { count++ }) }

// UseDirect: same interface, implementation in the victim's import tree.
func UseDirect(cur realm) {
	var r lendlib.Runner = lendlib.Direct{}
	r.Run(func() { count++ })
}

// UseOther lends its literal to another realm's crossing call.
func UseOther(cur realm) { lendother.Run(cross(cur), func() { count++ }) }

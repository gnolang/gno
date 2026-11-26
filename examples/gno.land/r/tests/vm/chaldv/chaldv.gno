// Package chaldv is a victim realm whose own literals are lent out.
package chaldv

import (
	lib "gno.land/p/demo/tests/chaldlib"
	"gno.land/r/tests/vm/chaldr"
)

var Count int

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

func Report(s string, err any) { println(s, Count, err) }

func Reenter(cur realm, g func()) any { return try(func() { g() }) }

// Attacks runs every attempt as the victim, in one crossing call.
func Attacks(cur realm) {
	Report("x01 other realm non-crossing", try(func() { chaldr.Call(func() { Count++ }) }))
	Report("x02 other realm crossing", chaldr.CallX(cross(cur), func() { Count++ }))
	Report("x03 other realm via lib", chaldr.ViaLib(cross(cur), func() { Count++ }))
	Report("x04 other realm keeps", chaldr.Keep(cross(cur), func() { Count++ }))
	Report("x05 other realm runs kept", chaldr.Run(cross(cur)))
	Report("x06 bounce back crossing", chaldr.Bounce(cross(cur), func() { Count++ }, Reenter))
	Report("x07 lib callback re-enters", try(func() {
		lib.CallThenCB(func() { Count++ }, func() { lib.Call(func() { Count++ }) })
	}))
	Report("x08 view victim lends", try(func() { lendFromView() }))
}

func lendFromView() { lib.Call(func() { Count++ }) }

// ViewLend is entered by others without crossing.
func ViewLend() { lib.Call(func() { Count++ }) }

func Kept() func() { return chaldr.Box[0] }

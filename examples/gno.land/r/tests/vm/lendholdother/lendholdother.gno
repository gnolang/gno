// Package lendholdother is another realm that receives lent literals; its crossing
// functions recover and return what the attempt panicked with.
package lendholdother

import lib "gno.land/p/demo/tests/lendholdlib"

var Box []func()

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

func Call(f func())                  { f() }
func CallX(cur realm, f func()) any  { return try(func() { f() }) }
func ViaLib(cur realm, f func()) any { return try(func() { lib.Call(f) }) }
func Keep(cur realm, f func()) any   { Box = append(Box, f); return nil }
func Run(cur realm) any              { return try(func() { Box[0]() }) }

// Bounce hands f back to the caller's crossing Reenter while still lent.
func Bounce(cur realm, f func(), back func(cur realm, g func()) any) any {
	return back(cross(cur), f)
}

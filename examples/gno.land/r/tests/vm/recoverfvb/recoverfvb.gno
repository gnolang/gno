// Package recoverfvb is a realm that keeps and makes closures, for the
// recover_fv file tests.
package recoverfvb

import "gno.land/p/demo/tests/recoverfv"

var (
	n    int
	kept func() any
	own  = recoverfv.New()
	// H is a library object this realm stores.
	H = &recoverfv.H{}
	// saved panics on this realm's object; savedInc writes this realm.
	saved    = func() { own.Set("x") }
	savedInc = func() { n++ }
)

func Keep(cur realm, f func() any) { kept = f }
func RunKept() any                 { return kept() }
func Saved() func()                { return saved }
func SavedInc() func()             { return savedInc }
func N() int                       { return n }

// Package entryclosure hands out closures over its state: some built by a
// library, one written in this realm's source.
package entryclosure

import "gno.land/p/demo/tests/entryclosure/v0"

var (
	box     = &entryclosure.Box{}
	counter = entryclosure.Counter()
	count   int
)

// Incer returns a library-built closure over box.
func Incer() func() { return box.Incer() }

// LitIncer returns a closure written in this realm's source.
func LitIncer() func() { return func() { box.N++ } }

// Counter returns the stored library-built closure over a captured local.
func Counter() func() int { return counter }

func N() int { return box.N }

func Count() int { return count }

// OwnerInc calls both library-built closures from this realm's own entry.
func OwnerInc(cur realm) {
	box.Incer()()
	count = counter()
}

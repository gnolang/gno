// Package entryclosure builds closures in library code, for tests of who a
// closure acts for when another realm calls it.
package entryclosure

type Box struct{ N int }

// Incer returns a closure that increments b.N.
func (b *Box) Incer() func() { return func() { b.N++ } }

// Counter returns a closure that increments a captured local.
func Counter() func() int {
	c := 0
	return func() int {
		c++
		return c
	}
}

package pfuncbind

import "gno.land/p/demo/tests/launderpkg/v0"

type Fn = func(*launderpkg.Immutable)

// Bind returns a closure that calls f(x).
func Bind(f func(Fn), x Fn) func() { return func() { f(x) } }

// Apply runs f on x.
func Apply(f Fn, x *launderpkg.Immutable) { f(x) }

// Clobber is the package's own exported writer.
func Clobber(x *launderpkg.Immutable) { x.Field = "pwnd-same-pkg" }

// Hook is a function type with a method that calls its receiver.
type Hook func(*launderpkg.Immutable)

func (h Hook) Do(x *launderpkg.Immutable) { h(x) }

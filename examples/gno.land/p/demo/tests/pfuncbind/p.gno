package pfuncbind

import "gno.land/p/demo/tests/launderpkg/v0"

type Fn = func(*launderpkg.Immutable)

// Bind returns a closure that calls f(x).
func Bind(f func(Fn), x Fn) func() { return func() { f(x) } }

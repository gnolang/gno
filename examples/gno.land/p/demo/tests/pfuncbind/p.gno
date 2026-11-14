package pfuncbind

import "gno.land/p/demo/tests/launderpkg/v0"

type Fn = func(*launderpkg.Immutable)

// Bind returns a closure that calls f(x).
func Bind(f func(Fn), x Fn) func() { return func() { f(x) } }

// Apply runs f on x.
func Apply(f Fn, x *launderpkg.Immutable) { f(x) }

// Clobber is the package's own exported writer.
func Clobber(x *launderpkg.Immutable) { x.Field = "pwnd-same-pkg" }

// Evil and EvilS write the object a caller's interface hands them.
type Evil int

func (Evil) Run(x *launderpkg.Immutable) { x.Field = "pwnd-via-iface" }

type EvilS struct{}

func (*EvilS) Run(x *launderpkg.Immutable) { x.Field = "pwnd-via-iface" }

type EvilV struct{ N int }

func (EvilV) Run(x *launderpkg.Immutable) { x.Field = "pwnd-via-iface" }

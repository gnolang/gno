// Package ifaceattack holds library types whose methods write the object a
// caller's interface hands them. Victim realms in tests must not import it.
package ifaceattack

import "gno.land/p/demo/tests/launderpkg/v0"

type Evil int

func (Evil) Run(x *launderpkg.Immutable) { x.Field = "pwnd-via-iface" }

type EvilS struct{}

func (*EvilS) Run(x *launderpkg.Immutable) { x.Field = "pwnd-via-iface" }

type EvilV struct{ N int }

func (EvilV) Run(x *launderpkg.Immutable) { x.Field = "pwnd-via-iface" }

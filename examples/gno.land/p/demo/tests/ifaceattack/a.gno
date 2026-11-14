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

// MakeH returns a method value; built in a view-only frame it has no stamp.
func MakeH() func(*launderpkg.Immutable) { return Evil(0).Run }

// Gen hands out such a method value through an interface method.
type Gen struct{}

func (Gen) Handler() func(*launderpkg.Immutable) { return Evil(0).Run }

// MakeSet takes a method value of t's Set; built view-only it has no stamp.
func MakeSet(t interface{ Set(string, any) bool }) func(string, any) bool { return t.Set }

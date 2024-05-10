package event_outer

import (
	"std"

	inner "gno.land/r/r3v4/event_inner"
)

func Event() {
	std.Emit("outer", "k1", "v1")
	inner.EventInner()
	std.Emit("outer", "k2", "v2")
}

func EventOuterOther() {
	std.Emit("outer_other", "k3", "v3")
	std.Emit("outer_other", "k4", "v4")
}

package event_inner

import (
	"std"
)

func EventInner() {
	std.Emit("inner", "k1", "v1")
	std.Emit("inner", "k2", "v2")
}

func EventInnerOther() {
	std.Emit("inner_other", "k3", "v3")
	std.Emit("inner_other", "k4", "v4")
}

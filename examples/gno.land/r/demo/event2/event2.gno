package event2

import (
	"std"
)

func Hello() {
	inner()
	std.Emit("t", "k", "v")
}

func inner() {
	std.Emit("t", "k", "v")
}

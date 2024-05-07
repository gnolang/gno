package event

import (
	"std"

	e2 "gno.land/r/demo/event2"
)

func Hello() {
	inner()
	std.Emit("t", "k", "v")
}

func inner() {
	e2.Hello()
	std.Emit("t", "k", "v")
}

func Other() {
	std.Emit("other", "k", "v")
}

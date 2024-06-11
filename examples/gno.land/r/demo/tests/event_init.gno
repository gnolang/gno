package tests

import (
	"std"
)

func init() {
	std.Emit("Foo",
		"bar", "42",
		"baz", "uwu",
	)
}

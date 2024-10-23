package event

import (
	"std"
)

func Emit(value string) {
	std.Emit("TAG", "key", value)
}

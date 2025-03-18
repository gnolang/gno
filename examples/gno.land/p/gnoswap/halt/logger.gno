package halt

import (
	"std"
)

const (
	errHalted           = "ERR_HALTED"
	errInvalidhaltLevel = "ERR_INVALID_HALT_LEVEL"
)

func Logger(event string, args ...string) {
	std.Emit(event, args...)
}

package halt

import (
	"std"
)

const (
	errInvalidhaltLevel = "ERR_INVALID_HALT_LEVEL"
)

func Logger(event string, args ...string) {
	std.Emit(event, args...)
}

package iavl

import (
	"fmt"
)

func debug(format string, args ...any) {
	if false {
		fmt.Printf(format, args...)
	}
}

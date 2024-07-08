package iavl

import (
	"fmt"
)

func debug(format string, args ...interface{}) {
	if false {
		fmt.Printf(format, args...)
	}
}

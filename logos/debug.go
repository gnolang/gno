package logos

import (
	"fmt"
)

// NOTE: the golang compiler doesn't seem to be intelligent
// enough to remove steps when const debug is True,
// so it is still faster to first check the truth value
// before calling debug.Println or debug.Printf.

const debug debugging = false // or flip

type debugging bool

func (d debugging) Println(args ...interface{}) {
	if d {
		fmt.Println(append([]interface{}{"DEBUG:"}, args...)...)
	}
}

func (d debugging) Printf(format string, args ...interface{}) {
	if d {
		fmt.Printf("DEBUG: "+format, args...)
	}
}

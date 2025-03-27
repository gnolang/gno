package complexargs

import (
	"strings"
)

var (
	slice []int
)

// SetSlice takes in a complex argument, which needs to be called via MsgRun
func SetSlice(newSlice []int) {
	slice = newSlice
}

func Render(_ string) string {
	out := "# Complex argument functions\n\n"
	out += `Exposed realm functions that take in complex arguments, such as slices, structs, pointers, etc,
cannot be called via a standard "MsgCall" transaction. To call these functions, users need to use "MsgRun".

`

	out = strings.Replace(out, "\"", "`", -1)
	return out
}

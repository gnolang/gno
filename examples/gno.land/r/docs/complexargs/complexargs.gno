package complexargs

import (
	"strings"
)

// Some example complex types, ie a slice, and a custom type
var (
	slice      []int
	coolObject CustomType
)

type CustomType struct {
	name    string
	numbers []int
}

// SetSlice takes in a complex argument - a slice - which needs to be called via MsgRun
func SetSlice(newSlice []int) {
	slice = newSlice
}

// Sets a new object
func SetCoolObject(newCoolObject CustomType) {
	coolObject = newCoolObject
}

func GetSlice() []int {
	return slice
}

func GetCoolObject() *CustomType {
	return &coolObject
}

func Render(_ string) string {
	out := "# Complex argument functions\n\n"
	out += `Exposed realm functions that take in complex arguments, such as slices, structs, pointers, etc,
cannot be called via a standard "MsgCall" transaction. To call these functions, users need to use "MsgRun".

`
	out = strings.Replace(out, "\"", "`", -1)
	return out
}

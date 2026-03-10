package complexargs

import (
	"strings"

	"gno.land/p/nt/ufmt/v0"
)

// Some example complex types, ie a slice, and a custom type
var (
	slice    = []int{1, 2, 3}
	myObject = &CustomType{
		Name:    "Alice",
		Numbers: []int{4, 5, 6},
	}
)

type CustomType struct {
	Name    string
	Numbers []int
}

// SetSlice takes a complex argument that must be called using MsgRun or from another contract that imports this one.
func SetSlice(_ realm, newSlice []int) {
	slice = newSlice
}

// SetMyObject takes a complex argument that must be called using MsgRun or from another contract that imports this one.
func SetMyObject(_ realm, newCoolObject CustomType) {
	myObject = &newCoolObject
}

func Render(_ string) string {
	out := "# Complex argument functions\n\n"
	out += `Exposed realm functions and methods that take in complex arguments, such as slices, structs, pointers, etc,
cannot be called via a standard "MsgCall" transaction. To call these functions, users need to use "MsgRun".

Check out the source code to see example functions and objects.

In this case, the following "MsgRun" code would be used to call the function:
`
	out += ufmt.Sprintf("```go\n%s\n```\n\n", msgrun)

	out += "Value of int slice: "
	for i, v := range slice {
		if i > 0 {
			out += ", "
		}
		out += ufmt.Sprintf("%d", v)
	}
	out += "\n\n"

	if myObject != nil {
		s := ""
		for i, v := range myObject.Numbers {
			if i > 0 {
				s += ","
			}
			s += ufmt.Sprintf("%d", v)
		}
		out += ufmt.Sprintf("Value of myObject: `CustomObject{Name: %s, Numbers: %s}`", myObject.Name, s)
	}

	out = strings.ReplaceAll(out, "\"", "`")
	return out
}

const msgrun = `package main

// Import the realm you want to call
import "gno.land/r/docs/complexargs"

func main() {
	// Create the complex arguments to pass:
	slice := []int{1, 2, 3}
	// Call the function
	complexargs.SetSlice(cross, slice)

	// The same can be done with custom types:
	obj := complexargs.CustomType{Name: "whatever", Numbers: []int{1, 10, 100}}
	complexargs.SetMyObject(obj)
}`

package complexargs

import (
	"strings"

	"gno.land/p/demo/ufmt"
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
func SetSlice(newSlice []int) {
	slice = newSlice
}

// SetMyObject takes a complex argument that must be called using MsgRun or from another contract that imports this one.
func SetMyObject(newCoolObject CustomType) {
	myObject = &newCoolObject
}

func Render(_ string) string {
	out := "# Complex argument functions\n\n"
	out += `Exposed realm functions that take in complex arguments, such as slices, structs, pointers, etc,
cannot be called via a standard "MsgCall" transaction. To call these functions, users need to use "MsgRun".

Check out the source code to see example functions and objects.

In this case, the following "MsgRun" code would be used to call the function:  
`
	out += ufmt.Sprintf("```go\n%s\n```\n\n", msgrun)

	out += "Value of int slice: "
	for _, v := range slice {
		out += ufmt.Sprintf("%d", v)
	}

	out += "\n\n"

	if myObject != nil {
		s := ""
		for _, v := range myObject.Numbers {
			s += ufmt.Sprintf("%d", v)
		}
		out += ufmt.Sprintf("Value of myObject: `CustomObject{Name: %s, Numbers: %s}`", myObject.Name, s)
	}

	out = strings.Replace(out, "\"", "`", -1)
	return out
}

const msgrun = `package main

// Import the realm you want to call
import "gno.land/r/docs/complexargs"

func main() {
	// Create the complex arguments to pass:
	slice := []int{1, 2, 3}
	// Call the function
	complexargs.SetSlice(slice)

	// The same can be done with custom types:
	obj := complexargs.CustomType{Name: "whatever", Numbers: slice}
	complexargs.SetMyObject(obj)
}`

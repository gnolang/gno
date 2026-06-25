package complexargs

import (
	"gno.land/p/nt/markdown/sanitize/v0"
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

// NewCustomType constructs a CustomType. Because the storage = authority rule
// forbids allocating a foreign realm's /r/-declared type outside its home
// realm, callers (including MsgRun ephemerals) must obtain CustomType values
// through this constructor rather than composing the struct literal directly.
func NewCustomType(name string, numbers []int) CustomType {
	return CustomType{Name: name, Numbers: numbers}
}

// SetSlice takes a complex argument that must be called using MsgRun or from
// another contract that imports this one. It is intentionally unguarded so the
// demo stays focused on argument passing; a production crossing function that
// derives caller identity must first check cur.IsCurrent(), then use
// cur.Previous() — see gno.land/r/docs/private.
func SetSlice(_ realm, newSlice []int) {
	slice = newSlice
}

// SetMyObject takes a complex argument that must be called using MsgRun or from
// another contract that imports this one. Like SetSlice, it is intentionally
// unguarded; gate equivalent setters with cur.IsCurrent() in real realms.
func SetMyObject(_ realm, newCoolObject CustomType) {
	myObject = &newCoolObject
}

func Render(_ string) string {
	out := "# Complex argument functions\n\n"
	out += "Exposed realm functions and methods that take in complex arguments, such as slices, structs, pointers, etc, " +
		"cannot be called via a standard `MsgCall` transaction. To call these functions, users need to use `MsgRun`.\n\n"
	out += "Check out the source code to see example functions and objects.\n\n"
	out += "Note that a realm's own struct type cannot be constructed outside that realm, so the realm exposes a constructor (`NewCustomType`) that `MsgRun` callers use to build the value before passing it back.\n\n"
	out += "In this case, the following `MsgRun` code would be used to call the function:\n"

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
		// myObject.Name is settable via SetMyObject, so it is untrusted input
		// echoed into markdown — sanitize it before output.
		out += ufmt.Sprintf("Value of myObject: `CustomType{Name: %s, Numbers: %s}`", sanitize.InlineText(myObject.Name), s)
	}

	return out
}

const msgrun = `package main

// Import the realm you want to call
import "gno.land/r/docs/complexargs"

func main(cur realm) {
	// Slices are anonymous types, so build and pass them directly.
	slice := []int{1, 2, 3}
	// Crossing functions are called with cross(cur).
	complexargs.SetSlice(cross(cur), slice)

	// A realm's own struct type cannot be constructed outside that realm
	// (the storage = authority rule), so obtain it from the realm's
	// constructor before passing it back.
	obj := complexargs.NewCustomType("whatever", []int{1, 10, 100})
	complexargs.SetMyObject(cross(cur), obj)
}`

// PKGPATH: gno.land/r/docs/complexargs/z_filetest

package z_filetest

import "gno.land/r/docs/complexargs"

// Demonstrates the MsgRun pattern the realm documents: complex arguments
// (slice, struct) passed into crossing functions via cross(cur).
func main(cur realm) {
	complexargs.SetSlice(cross(cur), []int{7, 8, 9})
	obj := complexargs.NewCustomType("Bob", []int{10, 20})
	complexargs.SetMyObject(cross(cur), obj)
	println(complexargs.Render(""))
}

// Output:
// # Complex argument functions
//
// Exposed realm functions and methods that take in complex arguments, such as slices, structs, pointers, etc, cannot be called via a standard `MsgCall` transaction. To call these functions, users need to use `MsgRun`.
//
// Check out the source code to see example functions and objects.
//
// Note that a realm's own struct type cannot be constructed outside that realm, so the realm exposes a constructor (`NewCustomType`) that `MsgRun` callers use to build the value before passing it back.
//
// In this case, the following `MsgRun` code would be used to call the function:
// ```go
// package main
//
// // Import the realm you want to call
// import "gno.land/r/docs/complexargs"
//
// func main(cur realm) {
// 	// Slices are anonymous types, so build and pass them directly.
// 	slice := []int{1, 2, 3}
// 	// Crossing functions are called with cross(cur).
// 	complexargs.SetSlice(cross(cur), slice)
//
// 	// A realm's own struct type cannot be constructed outside that realm
// 	// (the storage = authority rule), so obtain it from the realm's
// 	// constructor before passing it back.
// 	obj := complexargs.NewCustomType("whatever", []int{1, 10, 100})
// 	complexargs.SetMyObject(cross(cur), obj)
// }
// ```
//
// Value of int slice: 7, 8, 9
//
// Value of myObject: `CustomType{Name: Bob, Numbers: 10,20}`

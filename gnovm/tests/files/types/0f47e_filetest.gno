package main

import "fmt"

type integer interface{}

// note, cast untyped nil to interface{}, hence comparable, special case
// ---BinaryExpr---, OP: EQL, lx: (const (undefined)), rx: (const (undefined)), lt: interface{}, rt: <nil>, isLeftConstExpr: true, isRightConstExpr true, isLeftUntyped: false, isRightUntyped: false
func main() {
	//println(integer(nil) == nil)
	println(integer(nil) == nil)
	fmt.Printf("%T \n", integer(nil))
}

// Output:
// true
// <nil>

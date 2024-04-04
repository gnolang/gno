package main

import "fmt"

// note, cast untyped nil to interface{}, hence comparable, special case
// ---BinaryExpr---, OP: EQL, lx: (const (undefined)), rx: (const (undefined)), lt: interface{}, rt: <nil>, isLeftConstExpr: true, isRightConstExpr true, isLeftUntyped: false, isRightUntyped: false
func main() {
	println(interface{}(nil) == (*int)(nil)) // refer to convertConst and convertTo
	fmt.Printf("%T \n", interface{}(nil))
}

// Output:
// true
// <nil>

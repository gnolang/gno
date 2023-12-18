package main

import "fmt"

type myPointer *int

func main() {
	// NOTE: this is deprecated logic.
	// rhs nil is not converted to left, it's undefined
	// flow. left and right are both const:
	// DEBUG: ---BinaryExpr---, OP: EQL, lx: (const (nil main.myPointer)), rx: (const (undefined)), lt: main.myPointer, rt: <nil>, isLeftConstExpr: true, isRightConstExpr true, isLeftUntyped: false, isRightUntyped: false
	// the have same specificity of 0 (neither is primitive)
	// so the conversion dir is ->, it won't convert since dt is nil(type and value of nil are both nil), refer to convertConst logic
	// so it results in runtime: "isEql, lv: (nil main.myPointer), rv: (undefined), lvu: false, rvc: true"
	// so, as a special case, they are actually equal when: one is undefined, and one is defined with specific type(slice, map, pointer, func) and nil value
	// or, think, we should convert from type-nil(nil) to typed(nil)->typed, value is nil
	if myPointer(nil) == nil {
		fmt.Println("Pointer is nil")
	} else {
		fmt.Println("Pointer is not nil")
	}
}

// Output:
// Pointer is nil

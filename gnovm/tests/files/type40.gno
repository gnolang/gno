package main

type (
	// PrimitiveType
	Number  = int32
	Number2 = Number

	// PointerType
	Pointer  = *int32
	Pointer2 = Pointer

	// Interface
	Interface  = any
	Interface2 = Interface

	// S
	Struct  = struct{ Name string }
	Struct2 = Struct
)

func fNumber(n Number)       { println(n) }
func fPointer(p Pointer)     { println(*p) }
func fInterface(i Interface) { println(i) }
func fStruct(s Struct)       { println(s.Name) }

func main() {
	var n Number2 = 5
	fNumber(n)

	var num int32 = 6
	var p Pointer2 = &num
	fPointer(p)

	var i Interface2
	i = 7
	fInterface(i)

	var s Struct2 = Struct2{Name: "yo"}
	fStruct(s)
}

// Output:
// 5
// 6
// 7
// yo

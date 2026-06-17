package main

type T struct{ V int }

type C = T // alias to a defined type

type Named struct{ V int } // distinct defined type, same shape as T

func main() {
	// Embedding the alias C names the field "C", embedding T names it "T",
	// so the two structs are distinct even though C and T are the same type.
	var x, y interface{} = struct{ C }{}, struct{ T }{}
	if x == y {
		panic("FAIL: struct{C} == struct{T}")
	}

	// And distinct from embedding an unrelated defined type of the same shape.
	x, y = struct{ C }{}, struct{ Named }{}
	if x == y {
		panic("FAIL: struct{C} == struct{Named}")
	}

	// The embedded field is accessible by the alias name C, not T.
	s := struct{ C }{}
	s.C.V = 9
	println(s.C.V)
}

// Output:
// 9

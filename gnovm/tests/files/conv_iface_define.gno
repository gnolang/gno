package main

const a = 0

func kind(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	case nil:
		return "nil"
	}
	return "other"
}

func main() {
	// A name defined from a constant converted to an interface is the
	// interface, also in the if/switch clause blocks: a later assignment
	// of another dynamic type is accepted.
	switch i := (interface{})(a); i.(type) {
	case [0]string:
		println("array")
	default:
		println(kind(i))
		i = "s"
		println(kind(i))
	}
	if i := any(a); true {
		println(kind(i))
		i = nil
		println(kind(i))
	}
	j := interface{}(5)
	println(kind(j))
	j = "s"
	println(kind(j))
}

// Output:
// int
// string
// int
// nil
// int
// string

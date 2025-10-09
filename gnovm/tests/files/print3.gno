package main

func main() {
	nested := makeNested(100000)
	println(nested)
}

func makeNested(depth int) interface{} {
	if depth <= 0 {
		return 0
	}
	return []interface{}{makeNested(depth - 1)}
}

// Output:
// slice[(slice[(slice[(slice[(slice[(slice[(slice[(slice[(slice[(slice[(... []interface {})] []interface {})] []interface {})] []interface {})] []interface {})] []interface {})] []interface {})] []interface {})] []interface {})] []interface {})]

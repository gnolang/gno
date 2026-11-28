package main

// A type-switch `X` reading an enclosing loop variable reads the loop's
// variable, not a same-named package-level one.

var x interface{} = "outer"

func main() {
	for x := interface{}(1); x != nil; x = nil {
		switch x := x.(type) {
		case int:
			println("int", x)
		case string:
			println("string", x)
		}
	}
	println(x.(string))
}

// Output:
// int 1
// outer

package main

func main() {
	var i any = 3.14
	switch i.(type) {
	default:
		println("default")
	case string:
		println("string")
	case int:
		println("int")
	}
}

// Output:
// default

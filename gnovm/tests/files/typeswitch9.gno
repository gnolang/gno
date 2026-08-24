package main

func main() {
	var i any = 42
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
// int

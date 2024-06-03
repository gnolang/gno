package main

func main() {
	var i interface{} = "truc"

	switch i.(type) {
	case string:
		println("string")
		fallthrough
	default:
		println("unknown")
	}
}

// Error:
// cannot fallthrough in type switch

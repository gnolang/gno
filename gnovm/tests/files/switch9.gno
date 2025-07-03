package main

func main() {
	var i any = "truc"

	switch i.(type) {
	case string:
		println("string")
		fallthrough
	default:
		println("unknown")
	}
}

// Error:
// main/switch9.gno:9:3-14: cannot fallthrough in type switch

// TypeCheckError:
// main/switch9.gno:9:3: cannot fallthrough in type switch

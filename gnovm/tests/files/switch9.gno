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
// main/files/switch9.gno:9:3: cannot fallthrough in type switch

// TypeCheckError:
// main/files/switch9.gno:9:3: cannot fallthrough in type switch

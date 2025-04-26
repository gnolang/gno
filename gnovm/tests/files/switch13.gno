package main

func main() {
	var i any

	switch a := i.(type) {
	case string:
		println("string", a+" ok")
	case i:
		println("i is dummy")
	default:
		println("unknown")
	}
}

// Error:
// main/files/switch13.gno:8:0: i<VPBlock(2,0)> is not a type

// TypeCheckError:
// main/files/switch13.gno:9:7: i is not a type

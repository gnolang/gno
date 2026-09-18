package main

func main() {
	x := 99
	switch x {
	case 1:
		println("one")
	default:
		println("default")
		fallthrough
	}
}

// Error:
// main/switch44.gno:10:3-14: cannot fallthrough final case in switch

// TypeCheckError:
// main/switch44.gno:10:3: cannot fallthrough final case in switch

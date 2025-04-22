package main

func main() {
	x := 2
	switch x {
	case 1:
		println(1)
	case 2:
		println(2)
		fallthrough
	}
}

// Error:
// main/files/switch8b.gno:10:3: cannot fallthrough final case in switch

// TypeCheckError:
// main/files/switch8b.gno:10:3: cannot fallthrough final case in switch

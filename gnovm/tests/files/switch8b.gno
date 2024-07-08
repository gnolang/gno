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
// cannot fallthrough final case in switch

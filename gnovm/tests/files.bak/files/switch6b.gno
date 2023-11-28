package main

func main() {
	a := 0
	switch b := "foo"; {
	case a == 0:
		println(200)
		fallthrough // next case is executed even if a != 3
	case a == 3:
		println(100)
	default:
		println(a, b)
	}
}

// Output:
// 200
// 100

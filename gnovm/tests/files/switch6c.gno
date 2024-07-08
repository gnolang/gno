package main

func main() {
	a := 0
	switch b := "foo"; {
	case a == 0:
		println(200)
		fallthrough
	case a == 3:
		println(100)
		fallthrough // multiple fallthrough
	default:
		println(a, b)
	}
}

// Output:
// 200
// 100
// 0 foo

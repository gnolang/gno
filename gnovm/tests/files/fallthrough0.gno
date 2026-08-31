package main

func main() {
	x := 0
	switch 1 {
	case 1:
		x = 100
		defer func() {
			println("defer:", x)
		}()
		fallthrough
	case 2:
		x = 200
		println("case2:", x)
	}
	println("after switch:", x)
}

// Output:
// case2: 200
// after switch: 200
// defer: 200

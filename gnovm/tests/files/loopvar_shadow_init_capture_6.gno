package main

// A for-init reading its own name reads the nearest declaration of it, even
// when an enclosing loop declares the same name further out: an if-init, a
// switch-init, a type-switch variable, a range key, a const or a named
// result in between all shadow the loop's variable.

func main() {
	for x := 0; x < 1; x++ {
		if x := 5; x > 0 {
			for x := x + 1; x < 7; x++ {
				println("if", x)
			}
		}
		switch x := 5; {
		default:
			for x := x + 1; x < 7; x++ {
				println("switch", x)
			}
		}
		switch x := interface{}(5).(type) {
		case int:
			for x := x + 1; x < 7; x++ {
				println("typeswitch", x)
			}
		}
		for x := range []int{0, 0, 0} {
			for x := x; x < 1; x++ {
				println("range", x)
			}
		}
		{
			const x = 5
			for x := x + 1; x < 7; x++ {
				println("const", x)
			}
		}
		func() (x int) {
			x = 5
			for x := x + 1; x < 7; x++ {
				println("result", x)
			}
			return
		}()
	}
}

// Output:
// if 6
// switch 6
// typeswitch 6
// range 0
// const 6
// result 6

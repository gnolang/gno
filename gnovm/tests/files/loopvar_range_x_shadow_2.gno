package main

// A range `X` reading an enclosing loop variable reads the loop's variable,
// not a same-named local or package-level one.

var y = 200

func main() {
	x := 100
	for x := 0; x < 1; x++ {
		for _, x := range []int{x} {
			println("local", x)
		}
	}
	for y := 0; y < 1; y++ {
		for _, y := range []int{y} {
			println("package", y)
		}
	}
	println(x, y)
}

// Output:
// local 0
// package 0
// 100 200

package main

type (
	BasicFunc func(int, int) int
	MyFunc    BasicFunc
)

func (f MyFunc) Apply(a, b int) int {
	return f(a, b)
}

func main() {
	basicAdd := func(a, b int) int {
		return a + b
	}
	var myAdd MyFunc
	myAdd = basicAdd

	result := myAdd.Apply(2, 3)
	println(result)
}

// Output:
// 5

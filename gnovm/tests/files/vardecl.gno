package main

func main() {
	var i interface{} = 1
	a, ok := i.(int)
	println(a, ok)

	b, c := doSomething()
	println(b, c)

	m := map[string]int{"a": 1, "b": 2}
	d, okk := m["d"]
	println(d, okk)
}

func doSomething() (int, string) {
	return 4, "hi"
}

// Output:
// 1 true
// 4 hi
// 0 false

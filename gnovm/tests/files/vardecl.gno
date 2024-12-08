package main

func main() {
	var i interface{} = 1
	var a, ok = i.(int)
	println(a, ok)

	var b, c = doSomething()
	println(b, c)

	m := map[string]int{"a": 1, "b": 2}
	var d, okk = m["d"]
	println(d, okk)
}

func doSomething() (int, string) {
	return 4, "hi"
}

// Output:
// 1 true
// 4 hi
// 0 false

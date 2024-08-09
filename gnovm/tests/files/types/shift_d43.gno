package main

func main() {
	a := 2
	s := map[string]interface{}{"k": 1 << a}
	println(s["k"])
}

// Output:
// 4

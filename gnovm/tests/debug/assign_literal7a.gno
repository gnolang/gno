package main

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	m["a"] = 0
	println(m)
}

// Output:
// map{("a" string):(0 int),("b" string):(2 int),("c" string):(3 int)}

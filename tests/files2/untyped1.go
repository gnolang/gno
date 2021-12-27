package main

func main() {
	a := 1
	b := 1
	m := map[bool]struct{}{}
	m[a == b] = struct{}{}
	println(m)
}

// Output:
// map{(true bool):(struct{} struct{})}

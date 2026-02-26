package main

func main() {
	var arr = [3]int{1, 2, 3}
	p := &arr
	s := p[0:1]
	println(s)
}

// Output:
// slice[(1 int)]
package main

func main() {
	type S struct {
		a float32
	}
	s := S{
		a: 1 << 2,
	}
	println(s.a)
}

// Output:
// 4

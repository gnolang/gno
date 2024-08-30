package main

type S struct {
	i int
}

func main() {
	println(&new(S).i)
}

// Output:
// &(0 int)

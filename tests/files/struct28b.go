package main

type T1 struct {
	T2
}

type T2 struct {
	*T1
}

func main() {
	t := T1{}
	println(t)
}

// Output:
// struct{(undefined)}

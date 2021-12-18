package main

type T1 struct {
	P []*T
}

type T2 struct {
	P2 *T
}

type T struct {
	*T1
	S1 *T
}

func main() {
	println(T2{})
}

// Output:
// struct{(nil *main.T)}

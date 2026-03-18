package main

type Inner struct {
	X int
	Y []int
}

type Outer struct {
	A Inner
	B string
}

func main() {
	_ = map[Outer]int{Outer{Inner{1, []int{2}}, "a"}: 10}
}

// Error:
// main/maplit14.gno:14:6-55: invalid map key type main.Outer

// TypeCheckError:
// main/maplit14.gno:14:10: invalid map key type Outer

package main

type I interface {
	dummy()
}
type IS struct{}

func (iss IS) dummy() {
}

type S struct {
	x int
	I
}

func main() {
	m := make(map[S]string)
	m[S{0, IS{}}] = "test"
	println(m)
}

// Output:
// map{(struct{(0 int),(struct{} main.IS)} main.S):("test" string)}

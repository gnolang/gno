package main

type A struct {
	D int
	Next *B
}

type B struct {
	C int
	Next *A
}

func main() {
	a := A{
		D: 11,
	}
	b := B{
		C: 22,
		Next: &a,
	}
	a.Next = &b
	println(a)
}

// Output:
// (struct{(11 int),(&(struct{(22 int),(&(ref@0 main.A) *main.A)} main.B) *main.B)} main.A)

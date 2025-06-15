package main

type A struct {
	D    int
	Next *B
}

type B struct {
	C    int
	Next *A
}

func main() {
	a := A{
		D: 11,
	}
	b := B{
		C:    22,
		Next: &a,
	}
	a.Next = &b
	println(&a)
	println(a) // this will print a redundant copied a.
}

// Output:
// &(struct{(11 int),(&(struct{(22 int),(ref@0 *main.A)} main.B) *main.B)} main.A)
// (struct{(11 int),(&(struct{(22 int),(&(struct{(11 int),(ref@1 *main.B)} main.A) *main.A)} main.B) *main.B)} main.A)

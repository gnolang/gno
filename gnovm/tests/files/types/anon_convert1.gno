package main

var MyFooPointerArray [3]*Foo

func init() {

	MyFooPointerArray = [3]*Foo{
		&Foo{A: 1, B: &Foo{A: 2}},
	}
}

type Foo struct {
	A int
	B *Foo
}

func main() {
	x := ([3]struct {
		A int
		B *Foo
	})(MyFooPointerArray) // <-- this should panic
	// cannot convert MyFooPointerArray (variable of type [3]*Foo) to type [3]struct{A int; B *Foo}
	println(x)

	// Output:
	// array[(&(struct{(1 int),(&(struct{(2 int),(nil *main.Foo)} main.Foo) *main.Foo)} main.Foo) *main.Foo),(nil *main.Foo),(nil *main.Foo)]

}

// Error:
// cannot convert MyFooPointerArray (variable of type [3]*Foo) to type [3]struct{A int; B *Foo}

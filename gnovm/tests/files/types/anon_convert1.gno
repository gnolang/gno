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
// main/anon_convert1.gno:18:7-21:23: cannot convert MyFooPointerArray<VPBlock(3,0)> (of type [3]*main.Foo) to type [3]struct{A int; B *main.Foo}

// TypeCheckError:
// main/anon_convert1.gno:21:5: cannot convert MyFooPointerArray (variable of type [3]*Foo) to type [3]struct{A int; B *Foo}

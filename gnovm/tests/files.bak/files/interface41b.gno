package main

type foo struct {
	bar string
}

func (f *foo) String() string {
	return "Hello from " + f.bar
}

type Stringer interface {
	String() string
}

func Foo(s string) Stringer {
	return &foo{s}
}

func main() {
	f := Foo("bar")
	println(f.String())
}

// Output:
// Hello from bar

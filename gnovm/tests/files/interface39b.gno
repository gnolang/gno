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

func main() {
	var f Stringer = &foo{bar: "bar"}
	println(*((*foo)(f)))
}

// Output:
// (struct{("bar" string)} main.foo)

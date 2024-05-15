package main

type foo interface {
	say()
}

func (f *foo) echo() int {
	return 1
}

type Bar struct{}

func (b *Bar) say() {}

func main() {
	var f *foo
	*f = &Bar{}
	println(f.echo())
}

// Error:
// main/files/type39b.gno:7: invalid receiver type *main.foo (base type is interface type)

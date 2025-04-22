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
// main/files/type39b.gno:7:1: invalid receiver type *main.foo (base type is interface type)

// TypeCheckError:
// main/files/type39b.gno:7:10: invalid receiver type foo (pointer or interface type); main/files/type39b.gno:18:12: f.echo undefined (type *foo is pointer to interface, not interface)

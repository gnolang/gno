package main

type foo interface {
	say()
}

type FF foo

func (f FF) echo() int {
	return 1
}

type Bar struct{}

func (b *Bar) say() {}

func main() {
	var f foo
	f = &Bar{}
	println(f.echo())
}

// Error:
// main/type39a.gno:9:1-11:2: invalid receiver type main.FF (base type is interface type)

// TypeCheckError:
// main/type39a.gno:9:9: invalid receiver type FF (pointer or interface type); main/type39a.gno:20:12: f.echo undefined (type foo has no field or method echo)

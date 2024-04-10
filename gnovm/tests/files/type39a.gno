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
// main/debug/type39a.gno:9: invalid receiver type main.FF (interface type)

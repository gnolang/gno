package main

type foo struct {
	name string
}

func makeFoo(name string) (foo, bool) {
	return foo{name}, true
}

func trymakeFoo(name string) func() {
	if f, ok := makeFoo(name); ok {
		return func() {
			println(f.name)
		}
	}
	return func() { println("not foo...") }
}

func main() {
	trymakeFoo("foo")()
	println("---ok")
}

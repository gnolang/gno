package main

type T struct{}

func (len T) M() {}

func main() {
	T{}.M()
}

// TypeCheckError:
// main/shadow_builtin_recv.gno:5:7: builtin identifiers cannot be shadowed: len

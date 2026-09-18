package main

type T struct{}

func (len T) M() {}

func main() {
	T{}.M()
}

// Error:
// main/shadow_builtin_recv.gno:5:1-20: builtin identifiers cannot be shadowed: len

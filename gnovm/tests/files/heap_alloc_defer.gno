package main

type Foo struct {
	num int
	f   func()
}

func main() {
	s := []Foo{
		{
			num: 1,
			f:   func() { println("hello") },
		},
		{
			num: 2,
			f:   func() { println("hola") },
		},
	}

	// tt is heap defined every iteration,
	// different with for loopvar spec.
	for _, tt := range s {
		f := func() {
			println(tt.num)
		}
		f()
		defer func() {
			tt.f()
		}()
	}
}

// Output:
// 1
// 2
// hola
// hola

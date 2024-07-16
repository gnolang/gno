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

	for _, tt := range s {
		state := false
		//f := func() {
		//	println(tt.num)
		//}
		//f()
		defer func() {
			println(state)
			tt.f()
		}()
	}
}

// Output:
// false
// hola
// false
// hello

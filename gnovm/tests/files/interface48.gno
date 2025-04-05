package main

type Writer interface {
	Write([]byte) (int, error)
}

type Stringer interface {
	String() string
}

func main() {
	var x interface {
		Writer
		Stringer
	}
	var w Writer = Writer(x) // explicit conversion
	println(w)
}

// Output:
// undefined

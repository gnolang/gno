package main

type S struct {
}

func main() {
	x := (*S)(nil)
	println("x == x", x == x)
	println("x == nil", x == nil)
	println("nil == x", nil == x)

	var y any = (*S)(nil)
	println("y == y", y == y)
	println("y == nil", y == nil)
	println("nil == y", nil == y)

	y = nil
	println("y == y", y == y)
	println("y == nil", y == nil)
	println("nil == y", nil == y)
}

// Output:
// x == x true
// x == nil true
// nil == x true
// y == y true
// y == nil false
// nil == y false
// y == y true
// y == nil true
// nil == y true

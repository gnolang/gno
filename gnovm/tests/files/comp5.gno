package main

type S struct {
}

func main() {
	x := (*S)(nil)
	_ = x // keep the variable used; the test is about something else
	var y any = (*S)(nil)
	var znil any = nil

	println("y == znil", y == znil)
	println("znil == y", znil == y)

	y = nil
	println("y == znil", y == znil)
	println("znil == y", znil == y)
}

// Output:
// y == znil false
// znil == y false
// y == znil true
// znil == y true

package main

type S struct {
}

func main() {
	x := (*S)(nil)
	var y interface{} = (*S)(nil)
	var znil interface{} = nil

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

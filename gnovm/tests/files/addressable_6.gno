package main

type S struct {
	a int
}

type Alias []int

func main() {
	// Type assertions copy the value being asserted, so only pointers and
	// slices are addressable. Slices are addressable because a copy of a slice
	// maintains a reference to the same underlying array.
	var i interface{}
	i = []int{1}
	println(&i.([]int)[0])

	i = &S{}
	println(&i.(*S).a)

	i = Alias{4}
	println(&i.(Alias)[0])
}

// Output:
// &(1 int)
// &(0 int)
// &(4 int)

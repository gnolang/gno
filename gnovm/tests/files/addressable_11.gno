package main

func main() {
	var ii **int
	i := new(int)
	ii = &i
	println(&(*ii))
	println(&ii)
	println(i)
	println(ii)
	println(&i)

	j := new(int)
	println(&(*j))

	println(&(*getPtr()))

	derefTypeAssert()
}

func getPtr() *int {
	return new(int)
}

func derefTypeAssert() {
	var i any
	i = new(int)
	println(&(*(i.(*int))))
}

// Output:
// &(&(0 int) *int)
// &(&(&(0 int) *int) **int)
// &(0 int)
// &(&(0 int) *int)
// &(&(0 int) *int)
// &(0 int)
// &(0 int)
// &(0 int)

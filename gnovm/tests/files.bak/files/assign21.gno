package main

type thing int

func (t *thing) pass() (*thing, bool) {
	return t, true
}

func main() {
	a_ := thing(1)
	a := &a_
	a, ok := a.pass()
	println(*a, ok)
}

// Output:
// (1 main.thing) true

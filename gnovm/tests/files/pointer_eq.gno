package main

type A *int

func main() {
	b := new(int)
	var a A
	a = b

	var c *int
	c = b

	println(areEqual(a, b))
	println(areEqual(b, c))
	println(CheckNil())
}

func areEqual(v, w interface{}) bool {
	return v == w
}

type Node struct{}

func CheckNil() bool {
	var n *Node
	return n.IsNil()
}

func (n *Node) IsNil() bool {
	return n == nil
}

// Output:
// false
// true
// true

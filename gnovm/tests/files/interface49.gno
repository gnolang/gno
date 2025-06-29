package main

func main() {
	var x A = aImpl{}
	x.B()
	x.C()
}

type A interface {
	interface {
		B()
	}

	C()
}

type aImpl struct{}

func (aImpl) B() { println(123) }
func (aImpl) C() { println(456) }

// Output:
// 123
// 456

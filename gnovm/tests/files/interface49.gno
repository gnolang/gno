// You can edit this code!
// Click here and start typing.
package main

import "fmt"

type D interface {
	B()
}
type A interface {
	C()
}

type (
	aImpl2 struct{}
	aImpl  struct {
		D
		aImpl2
	}
)

func (aImpl) B() { println(123) }
func (aImpl) C() { println(456) }
func main() {
	fmt.Println("Hello, 世界")
}

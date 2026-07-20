// MAXALLOC: 500000000
package main

// Type switch interface case charges gas per interface method.
type I interface {
	M0()
}

type S struct{}

func (S) M0() {}

func main() {
	var x interface{} = S{}
	switch x.(type) {
	case I:
	}
}

// Gas:
// 5367

package main

import "fmt"

type IntArray []int
type Arr *IntArray

func (a Arr) Add(x int) { // receiver is val, not ptr
	*a = append(*a, x)
}

func main() {
	a := new(IntArray)
	Arr(a).Add(4)

	fmt.Println(*a)
}

// Error:
// main/type37.gno:8:1-10:2: invalid receiver type main.Arr (base type is pointer type)

// TypeCheckError:
// main/type37.gno:8:9: invalid receiver type Arr (pointer or interface type); main/type37.gno:14:9: Arr(a).Add undefined (type Arr has no field or method Add)

// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/ufmt"
)

type A struct {
	nums []int
}

var (
	intPtr        *int
	strPtr        *string
	aPtr          *A
	concretePtr   *int
	concreteValue int
)

func init() {
	New()
}

func main(cur realm) {
	Delta()
	println(Values())
}

func New() {
	intPtr = new(int)
	strPtr = new(string)
	aPtr = &A{}
	concretePtr = &concreteValue
}

func Delta() {
	*intPtr++
	*strPtr += "hello"
	*aPtr = A{nums: []int{8, 5, 8}}
	*concretePtr = 100
}

func Values() string {
	var results string
	results += ufmt.Sprintf("%d, %s, %d, %d", *intPtr, *strPtr, *concretePtr, concreteValue)
	for _, n := range aPtr.nums {
		results += ufmt.Sprintf(", %d", n)
	}

	return results
}

// Output:
// 1, hello, 100, 100, 8, 5, 8

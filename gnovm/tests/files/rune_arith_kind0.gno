package main

import "fmt"

// An operation between an untyped rune and an untyped int is an untyped rune.
var (
	r1 = 'a' + 1
	r2 = 1 + 'a'
	r3 = 'a' * 2
	r4 = 'a' << 1
	r5 = 'a' - 1
	r6 = 'a' % 7
	r7 = 1.5 + 'a'
	r  = []rune{r1, r2, r3, r4, r5, r6}
)

func main() {
	fmt.Printf("%v %T\n", r1, r1)
	fmt.Printf("%v %T\n", r2, r2)
	fmt.Printf("%v %T\n", r3, r3)
	fmt.Printf("%v %T\n", r4, r4)
	fmt.Printf("%v %T\n", r5, r5)
	fmt.Printf("%v %T\n", r6, r6)
	fmt.Printf("%v %T\n", r7, r7)
	fmt.Println(r)
	var x int64 = 'a' + 1
	println(x)
	var i interface{} = 'a' + 1
	_, ok := i.(rune)
	println(ok)
	const c = 'a' + 1
	var y rune = c
	println(y, string(c))
	// Beyond int32 the pair folds in untyped int: gc's value in a typed slot.
	var big int64 = 'a' + (1 << 40)
	println(big, string('A'+0x100000000))
}

// Output:
// 98 int32
// 98 int32
// 194 int32
// 194 int32
// 96 int32
// 6 int32
// 98.5 float64
// [98 98 194 194 96 6]
// 98
// true
// 98 b
// 1099511627873 �

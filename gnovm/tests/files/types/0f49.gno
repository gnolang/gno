package main

import (
	"fmt"
)

// implicit type conversion and equality
// NOTE:, the output type is incorrect, but the == eval is right
func main() {
	{ // case 1 with slices
		type X []byte
		var x interface{} = []byte("foo")
		var xx interface{} = X(x.([]byte))
		fmt.Printf("%T, %v \n", x, x)   // []uint8 [102 111 111]
		fmt.Printf("%T, %v \n", xx, xx) // main.X  [102 111 111]
		fmt.Println(x == xx)            // false
	}

	{ // case 2 with arrays
		type X [3]byte
		var x interface{} = [3]byte{102, 111, 111}
		var xx interface{} = X(x.([3]byte))
		fmt.Printf("%T, %v \n", x, x)   // [3]uint8 [102 111 111]
		fmt.Printf("%T, %v \n", xx, xx) // main.X [102 111 111]
		fmt.Println(x == xx)            // false still.
	}
	{ // switch without conversion
		type X [3]byte
		var x interface{} = [3]byte{102, 111, 111}
		switch x {
		case [3]byte{102, 111, 111}:
			fmt.Println("ok") // ok
		case [3]byte{102, 111, 112}:
			panic("should not happen")
		default:
			panic("should not happen")
		}
	}

	{ // case 1 switch with conversion no match
		type X [3]byte
		var x interface{} = X([3]byte{102, 111, 111})
		switch x {
		case [3]byte{102, 111, 111}:
			panic("should not happen")
		case [3]byte{102, 111, 112}:
			panic("should not happen")
		default:
			fmt.Println("no match") // no match
		}
	}

	{ // case 2 switch with conversion match
		type X [3]byte
		var x interface{} = X([3]byte{102, 111, 111})
		switch x {
		case [3]byte{102, 111, 111}:
			panic("should not happen")
		case X([3]byte{102, 111, 111}):
			fmt.Println("ok") // ok
		case X([3]byte{102, 111, 112}):
			panic("should not happen")
		default:
			panic("should not happen")
		}
	}
}

// Output:
// []uint8, [102 111 111]
// []uint8, [102 111 111]
// false
// [3]uint8, [102 111 111]
// [3]uint8, [102 111 111]
// false
// ok
// no match
// ok

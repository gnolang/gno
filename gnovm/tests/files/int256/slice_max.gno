package main

// A supplied 3-index max is always bounds-checked, also when it is -1 or a
// 256-bit value below MinInt64 (which an index conversion saturates to -1).
func main() {
	s := []int{1, 2}
	for _, n := range []int256{-1, -1 << 100, 1 << 100} {
		func() {
			defer func() { println("recovered:", recover()) }()
			println(len(s[:1:n]))
		}()
	}
	var m int = -1
	func() {
		defer func() { println("recovered:", recover()) }()
		println(len(s[:1:m]))
	}()
	println(len(s[:1:int256(2)]), cap(s[:1:uint256(1)]))
}

// Output:
// recovered: runtime error: invalid slice index -1 (index must be non-negative)
// recovered: runtime error: invalid slice index -1 (index must be non-negative)
// recovered: runtime error: slice bounds out of range [0:1:9223372036854775807] with capacity 2
// recovered: runtime error: invalid slice index -1 (index must be non-negative)
// 1 1

// TypeCheckError:
// main/slice_max.gno:7:33: cannot use -1 << 100 (untyped int constant -1267650600228229401496703205376) as int256 value in array or slice literal (overflows); main/slice_max.gno:7:44: cannot use 1 << 100 (untyped int constant 1267650600228229401496703205376) as int256 value in array or slice literal (overflows)

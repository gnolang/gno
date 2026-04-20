// PKGPATH: gno.land/r/inline_arr_nested
package inline_arr_nested

// Test: inline array nested in struct, init() mutation persists.

type Slot struct {
	prices [4]int
}

var slot = Slot{prices: [4]int{100, 200, 300, 400}}

func init() {
	slot.prices[0] = 111
}

func main(cur realm) {
	println(slot.prices[0])
	println(slot.prices[3])
}

// Output:
// 111
// 400

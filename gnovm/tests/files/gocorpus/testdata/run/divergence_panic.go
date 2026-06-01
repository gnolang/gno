// Verifies the symmetric triple on a substantive runtime divergence:
// recovering from an out-of-bounds slice access yields different
// panic-value wording in Gno vs Go. The triple at the bottom pins
// both sides explicitly so a reader sees the difference without
// running anything; the harness also verifies they still actually
// differ.

package main

import "fmt"

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r)
		}
	}()
	a := []int{1, 2, 3}
	_ = a[5]
}

// GnoOutput:
// recovered: runtime error: slice index out of bounds: 5 (len=3)

// GoOutput:
// recovered: runtime error: index out of range [5] with length 3

// KnownDivergence: error-wording: same kind of out-of-range panic, different wording in the recovered value.
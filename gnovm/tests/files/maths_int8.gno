package main

import "math"

func main() {
	ctr := int64(0)

	for a64 := int64(math.MinInt8); a64 <= int64(math.MaxInt8); a64++ {
		for b64 := int64(math.MinInt8); b64 <= int64(math.MaxInt8); b64++ {
			a8 := int8(a64)
			b8 := int8(b64)
			if int64(a8) != a64 || int64(b8) != b64 {
				panic("LOGIC FAILURE IN TEST")
			}
			ctr++

			// ADDITION
			{
				r64 := a64 + b64

				// now the verification
				result, ok := math.Add8(a8, b8)
				if int64(math.MinInt8) <= r64 && r64 <= int64(math.MaxInt8) {
					if !ok || int64(result) != r64 {
						println("add", a8, b8, result, r64)
						panic("incorrect result for non-overflow")
					}
				} else {
					if ok {
						panic("incorrect ok result")
					}
				}
			}

			// SUBTRACTION
			{
				r64 := a64 - b64

				// now the verification
				result, ok := math.Sub8(a8, b8)
				if int64(math.MinInt8) <= r64 && r64 <= int64(math.MaxInt8) {
					if !ok || int64(result) != r64 {
						println("sub", a8, b8, result, r64)
						panic("incorrect result for non-overflow")
					}
				} else {
					if ok {
						panic("incorrect ok result")
					}
				}
			}

			// MULTIPLICATION
			{
				r64 := a64 * b64

				// now the verification
				result, ok := math.Mul8(a8, b8)
				if int64(math.MinInt8) <= r64 && r64 <= int64(math.MaxInt8) {
					if !ok || int64(result) != r64 {
						println("mul", a8, b8, result, r64)
						panic("incorrect result for non-overflow")
					}
				} else {
					if ok {
						panic("incorrect ok result")
					}
				}
			}

			// DIVISION
			if b8 != 0 {
				r64 := a64 / b64

				// now the verification
				result, _, ok := math.Quo8(a8, b8)
				if int64(math.MinInt8) <= r64 && r64 <= int64(math.MaxInt8) {
					if !ok || int64(result) != r64 {
						panic("incorrect result for non-overflow")
					}
				} else {
					if ok {
						panic("incorrect ok result")
					}
				}
			}
		}
	}
	println("done", ctr)
}

// Output:
// done 65536

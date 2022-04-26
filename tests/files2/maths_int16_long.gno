package main

import (
	"gno.land/p/maths"
)

func main() {
	ctr := int64(0)

	for a64 := int64(maths.MinInt16); a64 <= int64(maths.MaxInt16); a64++ {
		for b64 := int64(maths.MinInt16); b64 <= int64(maths.MaxInt16); b64++ {
			a16 := int16(a64)
			b16 := int16(b64)
			if int64(a16) != a64 || int64(b16) != b64 {
				panic("LOGIC FAILURE IN TEST")
			}
			ctr++

			// ADDITION
			{
				r64 := a64 + b64

				// now the verification
				result, ok := maths.Add16(a16, b16)
				if int64(maths.MinInt16) <= r64 && r64 <= int64(maths.MaxInt16) {
					if !ok || int64(result) != r64 {
						println("add", a16, b16, result, r64)
						panic("incorrect result for non-overflow")
					}
				} else {
					if ok {
						println("add", a16, b16, result, r64)
						panic("incorrect ok result")
					}
				}
			}

			// SUBTRACTION
			{
				r64 := a64 - b64

				// now the verification
				result, ok := maths.Sub16(a16, b16)
				if int64(maths.MinInt16) <= r64 && r64 <= int64(maths.MaxInt16) {
					if !ok || int64(result) != r64 {
						println("sub", a16, b16, result, r64)
						panic("incorrect result for non-overflow")
					}
				} else {
					if ok {
						println("sub", a16, b16, result, r64)
						panic("incorrect ok result")
					}
				}
			}

			// MULTIPLICATION
			{
				r64 := a64 * b64

				// now the verification
				result, ok := maths.Mul16(a16, b16)
				if int64(maths.MinInt16) <= r64 && r64 <= int64(maths.MaxInt16) {
					if !ok || int64(result) != r64 {
						println("mul", a16, b16, result, r64)
						panic("incorrect result for non-overflow")
					}
				} else {
					if ok {
						println("mul", a16, b16, result, r64)
						panic("incorrect ok result")
					}
				}
			}

			// DIVISION
			if b16 != 0 {
				r64 := a64 / b64

				// now the verification
				result, _, ok := maths.Quo16(a16, b16)
				if int64(maths.MinInt16) <= r64 && r64 <= int64(maths.MaxInt16) {
					if !ok || int64(result) != r64 {
						println("quo", a16, b16, result, r64)
						panic("incorrect result for non-overflow")
					}
				} else {
					if ok {
						println("quo", a16, b16, result, r64)
						panic("incorrect ok result")
					}
				}
			}
		}
	}
	println("done", ctr)
}

// Output:
// done XXXXXX

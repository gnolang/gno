package overflow

import (
	"math"
	"testing"
)

// sample all possibilities of 8 bit numbers
// by checking against 64 bit numbers

func TestAlgorithmsInt(t *testing.T) {
	errors := 0

	for a64 := int64(math.MinInt8); a64 <= int64(math.MaxInt8); a64++ {
		for b64 := int64(math.MinInt8); b64 <= int64(math.MaxInt8) && errors < 10; b64++ {

			a8 := int8(a64)
			b8 := int8(b64)

			if int64(a8) != a64 || int64(b8) != b64 {
				t.Fatal("LOGIC FAILURE IN TEST")
			}

			// ADDITION
			{
				r64 := a64 + b64

				// now the verification
				result, ok := Add8(a8, b8)
				if ok && int64(result) != r64 {
					t.Errorf("failed to fail on %v + %v = %v instead of %v\n",
						a8, b8, result, r64)
					errors++
				}
				if !ok && int64(result) == r64 {
					t.Fail()
					errors++
				}
			}

			// SUBTRACTION
			{
				r64 := a64 - b64

				// now the verification
				result, ok := Sub8(a8, b8)
				if ok && int64(result) != r64 {
					t.Errorf("failed to fail on %v - %v = %v instead of %v\n",
						a8, b8, result, r64)
				}
				if !ok && int64(result) == r64 {
					t.Fail()
					errors++
				}
			}

			// MULTIPLICATION
			{
				r64 := a64 * b64

				// now the verification
				result, ok := Mul8(a8, b8)
				if ok && int64(result) != r64 {
					t.Errorf("failed to fail on %v * %v = %v instead of %v\n",
						a8, b8, result, r64)
					errors++
				}
				if !ok && int64(result) == r64 {
					t.Fail()
					errors++
				}
			}

			// DIVISION
			if b8 != 0 {
				r64 := a64 / b64
				rem64 := a64 % b64

				// now the verification
				result, rem, ok := Quo8(a8, b8)
				if ok && int64(result) != r64 {
					t.Errorf("failed to fail on %v / %v = %v instead of %v\n",
						a8, b8, result, r64)
					errors++
				}
				if ok && int64(rem) != rem64 {
					t.Errorf("failed to fail on %v %% %v = %v instead of %v\n",
						a8, b8, rem, rem64)
					errors++
				}
			}
		}
	}
}

func TestAlgorithmsUint(t *testing.T) {

	errors := 0

	for a64 := uint64(0); a64 <= uint64(math.MaxUint8); a64++ {

		for b64 := uint64(0); b64 <= uint64(math.MaxUint8) && errors < 10; b64++ {

			a8 := uint8(a64)
			b8 := uint8(b64)

			if uint64(a8) != a64 || uint64(b8) != b64 {
				t.Fatal("LOGIC FAILURE IN TEST")
			}

			// ADDITION
			{
				r64 := a64 + b64

				// now the verification
				result, ok := UAdd8(a8, b8)
				if ok && uint64(result) != r64 {
					t.Errorf("failed to fail on %v + %v = %v instead of %v\n",
						a8, b8, result, r64)
					errors++
				}
				if !ok && uint64(result) == r64 {
					t.Fail()
					errors++
				}
			}

			// SUBTRACTION
			{
				r64 := a64 - b64

				// now the verification
				result, ok := USub8(a8, b8)
				if ok && uint64(result) != r64 {
					t.Errorf("failed to fail on %v - %v = %v instead of %v\n",
						a8, b8, result, r64)
				}
				if !ok && uint64(result) == r64 {
					t.Fail()
					errors++
				}
			}

			// MULTIPLICATION
			{
				r64 := a64 * b64

				// now the verification
				result, ok := UMul8(a8, b8)
				if ok && uint64(result) != r64 {
					t.Errorf("failed to fail on %v * %v = %v instead of %v\n",
						a8, b8, result, r64)
					errors++
				}
				if !ok && uint64(result) == r64 {
					t.Fail()
					errors++
				}
			}

			// DIVISION
			if b8 != 0 {
				r64 := a64 / b64

				// now the verifiggcation
				result, _, ok := UQuo8(a8, b8)
				if ok && uint64(result) != r64 {
					t.Errorf("failed to fail on %v / %v = %v instead of %v\n",
						a8, b8, result, r64)
					errors++
				}
				if !ok && result != 0 && uint64(result) == r64 {
					t.Fail()
					errors++
				}
			}
		}
	}

}

func TestQuotient(t *testing.T) {
	q, r, ok := Quo(100, 3)
	if r != 1 || q != 33 || !ok {
		t.Errorf("expected 100/3 => 33, r=1")
	}
	if _, _, ok = Quo(1, 0); ok {
		t.Error("unexpected lack of failure")
	}
	if _, _, ok = Quo(math.MinInt, -1); ok {
		t.Error("unexpected lack of failure")
	}
	uq, ur, ok := UQuo(100, 3)
	if ur != 1 || uq != 33 || !ok {
		t.Errorf("expected 100/3 => 33, r=1")
	}
	if _, _, ok = UQuo(1, 0); ok {
		t.Error("unexpected lack of failure")
	}
}

func TestAbs(t *testing.T) {
	if _, ok := Abs(math.MinInt); ok {
		t.Error("unexpected ok, absolute value overflow")
	}
	if _, ok := Abs64(math.MinInt64); ok {
		t.Error("unexpected ok, absolute value overflow")
	}
	if _, ok := Abs32(math.MinInt32); ok {
		t.Error("unexpected ok, absolute value overflow")
	}
	if _, ok := Abs16(math.MinInt16); ok {
		t.Error("unexpected ok, absolute value overflow")
	}
	if _, ok := Abs8(math.MinInt8); ok {
		t.Error("unexpected ok, absolute value overflow")
	}
}

func TestLong(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctr := int64(0)

	for a64 := int64(math.MinInt16); a64 <= int64(math.MaxInt16); a64++ {
		for b64 := int64(math.MinInt16); b64 <= int64(math.MaxInt16); b64++ {
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
				result, ok := Add16(a16, b16)
				if int64(math.MinInt16) <= r64 && r64 <= int64(math.MaxInt16) {
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
				result, ok := Sub16(a16, b16)
				if int64(math.MinInt16) <= r64 && r64 <= int64(math.MaxInt16) {
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
				result, ok := Mul16(a16, b16)
				if int64(math.MinInt16) <= r64 && r64 <= int64(math.MaxInt16) {
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
				result, _, ok := Quo16(a16, b16)
				if int64(math.MinInt16) <= r64 && r64 <= int64(math.MaxInt16) {
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

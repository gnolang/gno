package overflow

import (
	"math"
	"testing"
)
import "fmt"

// sample all possibilities of 8 bit numbers
// by checking against 64 bit numbers

func TestAlgorithms(t *testing.T) {

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

				// now the verification
				result, _, ok := Quotient8(a8, b8)
				if ok && int64(result) != r64 {
					t.Errorf("failed to fail on %v / %v = %v instead of %v\n",
						a8, b8, result, r64)
					errors++
				}
				if !ok && result != 0 && int64(result) == r64 {
					t.Fail()
					errors++
				}
			}
		}
	}

}

func TestQuotient(t *testing.T) {
	q, r, ok := Quotient(100, 3)
	if r != 1 || q != 33 || !ok {
		t.Errorf("expected 100/3 => 33, r=1")
	}
	if _, _, ok = Quotient(1, 0); ok {
		t.Error("unexpected lack of failure")
	}
}

//func TestAdditionInt(t *testing.T) {
//	fmt.Printf("\nminint8 = %v\n", math.MinInt8)
//	fmt.Printf("maxint8 = %v\n\n", math.MaxInt8)
//	fmt.Printf("maxint32 = %v\n", math.MaxInt32)
//	fmt.Printf("minint32 = %v\n\n", math.MinInt32)
//	fmt.Printf("maxint64 = %v\n", math.MaxInt64)
//	fmt.Printf("minint64 = %v\n\n", math.MinInt64)
//}

func Test64(t *testing.T) {
	fmt.Println("64bit:", _is64Bit())
}

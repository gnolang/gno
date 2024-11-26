package gnolang

import (
	"math"
	"strings"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"
)

func TestConvertUntypedBigdecToFloat(t *testing.T) {
	t.Parallel()

	dst := &TypedValue{}

	dec, err := apd.New(-math.MaxInt64, -4).SetFloat64(math.SmallestNonzeroFloat64 / 2)
	require.NoError(t, err)
	bd := BigdecValue{
		V: dec,
	}

	typ := Float64Type

	ConvertUntypedBigdecTo(dst, bd, typ)

	require.Equal(t, float64(0), dst.GetFloat64())
}

func TestBitShiftingOverflow(t *testing.T) {
	t.Parallel()

	testFunc := func(source, msg string) {
		defer func() {
			if len(msg) == 0 {
				return
			}

			r := recover()

			if r == nil {
				t.Fail()
			}

			err := r.(*PreprocessError)
			c := strings.Contains(err.Error(), msg)
			if !c {
				t.Fatalf(`expected "%s", got "%s"`, msg, r)
			}
		}()

		m := NewMachine("test", nil)

		n := MustParseFile("main.go", source)
		m.RunFiles(n)
		m.RunMain()
	}

	type cases struct {
		source string
		msg    string
	}

	tests := []cases{
		{
			`package test

func main() {
	const a = int32(1) << 33
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test

func main() {
	const a1 = int8(1) << 8
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test

func main() {
	const a2 = int16(1) << 16
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test

func main() {
	const a3 = int32(1) << 33
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test

func main() {
	const a4 = int64(1) << 65
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test

func main() {
	const b1 = uint8(1) << 8
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test

func main() {
    const b2 = uint16(1) << 16
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test

func main() {
	const b3 = uint32(1) << 33
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test

func main() {
    const b4 = uint64(1) << 65
}`,
			`test/main.go:3:1: constant overflows`,
		},
		{
			`package test
		
		func main() {
			const c1 = 1 << 128
		}`,
			``,
		},
		{
			`package test
		
		func main() {
			const c1 = 1 << 128
			println(c1)
		}`,
			`test/main.go:5:4: bigint overflows target kind`,
		},
	}

	for _, tc := range tests {
		testFunc(tc.source, tc.msg)
	}
}

func TestSubUnderflow(t *testing.T) {
	t.Parallel()

	testFunc := func(source, msg string) {
		defer func() {
			if len(msg) == 0 {
				return
			}

			r := recover()

			if r == nil {
				t.Fail()
			}

			err := r.(*PreprocessError)
			c := strings.Contains(err.Error(), msg)
			if !c {
				t.Fatalf(`expected "%s", got "%s"`, msg, r)
			}
		}()

		m := NewMachine("test", nil)

		n := MustParseFile("main.go", source)
		m.RunFiles(n)
		m.RunMain()
	}

	type cases struct {
		source string
		msg    string
	}

	tests := []cases{
		{
			`package test
		
		func main() {
			const u1 = uint8(0) - 1
		}`,
			`test/main.go:4:15: constant underflow`,
		},
		{
			`package test
		
		func main() {
			const u1 = uint16(0) - 1
		}`,
			`test/main.go:4:15: constant underflow`,
		},
		{
			`package test
		
		func main() {
			const u1 = uint32(0) - 1
		}`,
			`test/main.go:4:15: constant underflow`,
		},
		{
			`package test
		
		func main() {
			const u1 = uint64(0) - 1
		}`,
			`test/main.go:4:15: constant underflow`,
		},
		{
			`package test
		
		func main() {
			const u1 = uint(0) - 1
		}`,
			`test/main.go:4:15: constant underflow`,
		},
		{
			`package test
		
		func main() {
			const u1 = int8(-128) - 1
		}`,
			`test/main.go:4:15: constant underflow`,
		},
		{
			`package test
		
		func main() {
		   const u1 = int16(-32768) - 1
		}`,
			`test/main.go:4:17: constant underflow`,
		},
		{
			`package test
		
		func main() {
			const u1 = int32(-2147483648) - 1
		}`,
			`test/main.go:4:15: constant underflow`,
		},
		{
			`package test
		
		func main() {
		   const u1 = int64(-9223372036854775808) - 1
		}`,
			`test/main.go:4:17: constant underflow`,
		},
		{
			`package test
		
		func main() {
			const u1 = int(-9223372036854775808) - 1
		}`,
			`test/main.go:4:15: constant underflow`,
		},
	}

	for _, tc := range tests {
		testFunc(tc.source, tc.msg)
	}
}

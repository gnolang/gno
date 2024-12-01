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

	require.Equal(t, ConvertToSoftFloat64(0), dst.GetFloat64())
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

package gnolang

import (
	"math"
	"testing"

	"github.com/cockroachdb/apd"
	"github.com/stretchr/testify/require"
)

func TestConvertUntypedBigdecToFloat(t *testing.T) {
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

// TODO: Can't convert int to float.
// func TestVarDeclTypeConversionInt2Float(t *testing.T) {
// 	m := NewMachine("test", nil)
// 	c := `package test
// func main() {
// 	x := 10
// 	println(float64(x))
// }`
// 	n := MustParseFile("main.go", c)
// 	m.RunFiles(n)
// 	m.RunMain()

// 	assertOutput(t, c, "+1.000000e+001\n")
// }

func TestShortenVarDeclTypeConversion(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test
func main() {
	println(bool(nil))
}`
	n := MustParseFile("main.go", c)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	m.RunFiles(n)
	m.RunMain()

	if err := m.CheckEmpty(); err != nil {
		t.Error(err)
	}
}

type testConversion struct {
	name     string
	src      string
	expected string
}

func TestVariableTypeConversion(t *testing.T) {
	tests := []testConversion{
		{
			name: "float to int",
			src: `package test
			func main() {
				var x float64 = 10.5
				println(int(x))
			}`,
			expected: "10\n",
		},
		{
			name: "float to int with short declaration",
			src: `package test
			func main() {
				x := 10.5
				println(int(x))
			}`,
			expected: "10\n",
		},
		{
			name: "eval expr to int",
			src: `package test
			func main() {
				x := 10 / 3
				println(int(x))
			}`,
			expected: "3\n",
		},
	}

	for _, test := range tests {
		m := NewMachine("test", nil)
		n := MustParseFile("main.go", test.src)
		m.RunFiles(n)
		m.RunMain()
		assertOutput(t, test.src, test.expected)

		if err := m.CheckEmpty(); err != nil {
			t.Error(err)
		}
	}
}

type testConversionPanic struct {
	name string
	src  string
}

func TestVariableTypeConversionShouldPanic(t *testing.T) {
	tests := []testConversionPanic{
		{
			name: "convert untyped float constant to int",
			src: `package test
			func main() {
				println(int(10.5))
			}`,
		},
		{
			name: "convert IntKind to BoolKind",
			src: `package test
			func main() {
				println(bool(1))
			}`,
		},
		{
			name: "convert eval expr to bool",
			src: `package test
			func main() {
				println(bool(10 / 5))
			}`,
		},
		{
			name: "convert nil to bool",
			src: `package test
			func main() {
				println(bool(nil))
			}`,
		},
		{
			name: "convert string to int",
			src: `package test
			func main() {
				println(int("123"))
			}`,
		},
		{
			name: "convert bool to int",
			src: `package test
			func main() {
				println(int(true))
			}`,
		},
		{
			name: "convert slice to int",
			src: `package test
			func main() {
				println(int([]int{1, 2, 3}))
			}`,
		},
		{
			name: "convert map to int",
			src: `package test
			func main() {
				println(int(map[string]int{"one": 1}))
			}`,
		},
	}

	for _, test := range tests {
		m := NewMachine("test", nil)
		n := MustParseFile("main.go", test.src)
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		m.RunFiles(n)
		m.RunMain()

		if err := m.CheckEmpty(); err != nil {
			t.Error(err)
		}
	}
}

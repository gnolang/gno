package gnolang

import (
	"testing"
)

type uverseTestCases struct {
	name     string
	code     string
	expected string
}

func TestIssue1337PrintNilSliceAsUndefined(t *testing.T) {
	test := []uverseTestCases{
		{
			name: "print empty slice",
			code: `package test
			func main() {
				emptySlice1 := make([]int, 0)
				emptySlice2 := []int{}

				println(emptySlice1)
				println(emptySlice2)
			}`,
			expected: "slice[]\nslice[]\n",
		},
		{
			name: "nil slice",
			code: `package test
			func main() {
				println(nil)
			}`,
			expected: "undefined\n",
		},
		{
			name: "print empty string slice",
			code: `package test
			func main() {
				var a []string
				println(a)
			}`,
			expected: "(nil []string)\n",
		},
		{
			name: "print non-empty slice",
			code: `package test
			func main() {
				a := []string{"a", "b"}
				println(a)
			}`,
			expected: "slice[(\"a\" string),(\"b\" string)]\n",
		},
		{
			name: "print empty map",
			code: `package test
			func main() {
				var a map[string]string
				println(a)
			}`,
			expected: "(nil map[string]string)\n",
		},
		{
			name: "print non-empty map",
			code: `package test
			func main() {
				a := map[string]string{"a": "b"}
				println(a)
			}`,
			expected: "map{(\"a\" string):(\"b\" string)}\n",
		},
		{
			name: "print nil struct",
			code: `package test
			func main() {
				var a struct{}
				println(a)
			}`,
			expected: "struct{}\n",
		},
		{
			name: "print function",
			code: `package test
			func foo(a, b int) int {
				return a + b
			}
			func main() {
				println(foo(1, 3))
			}`,
			expected: "4\n",
		},
		{
			name: "print composite slice",
			code: `package test
			func main() {
				a, b, c, d := 1, 2, 3, 4
				x := []int{
					a: b,
					c: d,
				}
				println(x)
			}`,
			expected: "slice[(0 int),(2 int),(0 int),(4 int)]\n",
		},
		{
			name: "simple recover case",
			code: `package test

			func main() {
				defer func() { println("recover", recover()) }()
				println("simple panic")
			}`,
			expected: "simple panic\nrecover undefined\n",
		},
		{
			name: "nested recover",
			code: `package test

			func main() {
				defer func() { println("outer recover", recover()) }()
				defer func() { println("nested panic") }()
				println("simple panic")
			}`,
			expected: "simple panic\nnested panic\nouter recover undefined\n",
		},
		{
			name: "print non-nil function",
			code: `package test
			func f() int {
				return 1
			}

			func main() {
				g := f
				println(g)
			}`,
			expected: "f\n",
		},
		{
			name: "print primitive types",
			code: `package test
			func main() {
				println(1)
				println(1.1)
				println(true)
				println("hello")
			}`,
			expected: "1\n1.1\ntrue\nhello\n",
		},
	}

	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMachine("test", nil)
			n := MustParseFile("main.go", tc.code)
			m.RunFiles(n)
			m.RunMain()
			assertOutput(t, tc.code, tc.expected)
		})
	}
}

func TestGetCapacityNilValue(t *testing.T) {
	tests := []uverseTestCases{
		{
			name: "cap of nil array",
			code: `
package test

func main() {
	var arr [5]int
	println(cap(&arr))
	var nilArr *[5]int
	println(cap(nilArr))
}`,
			expected: "5\n5\n",
		},
		{
			name: "cap of nil slice",
			code: `
package test

func main() {
	var slice []int
	println(cap(slice))
}`,
			expected: "0\n",
		},
		{
			name: "cap of nil array in function",
			code: `
package test

func main() {
	printCap(nil)
}

func printCap(arr *[3]string) {
	println(cap(arr))
}`,
			expected: "3\n",
		},
		{
			name: "cap of different nil array types",
			code: `
package test

func main() {
	var nilIntArr *[4]int
	var nilFloatArr *[6]float64
	var nilStringArr *[2]string
	println(cap(nilIntArr))
	println(cap(nilFloatArr))
	println(cap(nilStringArr))
}`,
			expected: "4\n6\n2\n",
		},
		{
			name: "cap of nil multidimensional array",
			code: `
package test

func main() {
	var nilMultiArr *[2][3]int
	println(cap(nilMultiArr))
}`,
			expected: "2\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMachine("test", nil)
			n := MustParseFile("main.go", tc.code)
			m.RunFiles(n)
			m.RunMain()
			assertOutput(t, tc.code, tc.expected)
		})
	}
}

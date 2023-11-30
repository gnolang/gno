package gnolang

import (
	"testing"
)

type printlnTestCases struct {
	name     string
	code     string
	expected string
}

func TestIssue1337PrintNilSliceAsUndefined(t *testing.T) {
	test := []printlnTestCases{
		{
			name: "nil slice",
			code: `package test
			func main() {
				println(nil)
			}`,
			expected: "undefined\n",
		},
		{
			name: "print empty slice",
			code: `package test
			func main() {
				var a []string
				println(a)
			}`,
			expected: "undefined\n",
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
			name: "print recover with panic",
			code: `package test

			func main() {
				f()
			}
		
			func f() {
				defer func() { println("f recover", recover()) }()
				defer g()
				panic("wtf")
			}
		
			func g() {
				println("g recover", recover())
			}`,
			expected: "g recover wtf\nf recover undefined\n",
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
			name: "print function 2",
			code: `package test
			func f() func() {
				return nil
			}
			
			func main() {
				g := f()
				println(g)
				if g == nil {
					println("nil func")
				}
			}`,
			expected: "nil func()\nnil func\n",
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

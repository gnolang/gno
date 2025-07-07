package repl

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type step struct {
	Line   string
	Error  string
	Result string
}

var fixtures = []struct {
	Name      string
	CodeSteps []step
}{
	{
		Name: "Add new import",
		CodeSteps: []step{
			{
				Line:   "import \"fmt\"\nimport \"os\"",
				Result: "",
			},
			{
				Line:   "fmt.Println(\"hello\")",
				Result: "hello",
			},
		},
	},
	{
		Name: "Add new constant",
		CodeSteps: []step{
			{
				Line:   "const test2, test3 = \"test_string2\", \"test_string3\"",
				Result: "",
			},
			{
				Line:   "const test = \"test_string\"",
				Result: "",
			},
			{
				Line:   "println(test, test2, test3)",
				Result: "test_string test_string2 test_string3",
			},
		},
	},
	{
		Name: "Add struct and functions",
		CodeSteps: []step{
			{
				Line:   "type MyStruct struct { count int}",
				Result: "",
			},
			{
				Line:   "func (s *MyStruct) Add(){s.count++}",
				Result: "",
			},
			{
				Line:   "s := MyStruct{1}; s.Add(); println(s.count)",
				Result: "2",
			},
		},
	},
	{
		Name: "Add new var",
		CodeSteps: []step{
			{
				Line:   "var test2, test3 string = \"test_string2\", \"test_string3\"",
				Result: "",
			},
			{
				Line:   "var test int = 42",
				Result: "",
			},
			{
				Line:   "println(test, test2, test3)",
				Result: "42 test_string2 test_string3",
			},
		},
	},
	{
		Name: "Add new define",
		CodeSteps: []step{
			{
				Line:   "var test2, test3 string = \"test_string2\", \"test_string3\"",
				Result: "",
			},
			{
				Line:   "test := 42",
				Result: "",
			},
			{
				Line:   "println(test, test2, test3)",
				Result: "42 test_string2 test_string3",
			},
		},
	},
	{
		Name: "Re-assign",
		CodeSteps: []step{
			{
				Line:   "var test2, test3 string = \"test_string2\", \"test_string3\"",
				Result: "",
			},
			{
				Line:   "test2 = \"something_new\"",
				Result: "",
			},
			{
				Line:   "println(test2, test3)",
				Result: "something_new test_string3",
			},
		},
	},
	{
		Name: "Add wrong code",
		CodeSteps: []step{
			{
				Line:  "importasdasd",
				Error: "name importasdasd not declared",
			},
			{
				Line: "var a := 1",
				// we cannot check the entire error because it is different depending on the used Go version.
				Error: "<repl>:1:14: expected type, found ':='",
			},
		},
	},
	{
		Name: "Add function and use it",
		CodeSteps: []step{
			{
				Line:   "func sum(a,b int)int{return a+b}",
				Result: "",
			},
			{
				Line:   "import \"fmt\"",
				Result: "",
			},
			{
				Line:   "fmt.Println(sum(1,1))",
				Result: "2",
			},
		},
	},
	{
		Name: "All declarations at once",
		CodeSteps: []step{
			{
				Line:   "import \"fmt\"\nfunc sum(a,b int)int{return a+b}",
				Result: "",
			},
			{
				Line:   "fmt.Println(sum(1,1))",
				Result: "2",
			},
		},
	},
	{
		Name: "Fibonacci",
		CodeSteps: []step{
			{
				Line: `
				func fib(n int)int {
					if n < 2 {
						return n
					}
					return fib(n-2) + fib(n-1)
				}
				`,
				Result: "",
			},
			{
				Line:   "println(fib(24))",
				Result: "46368",
			},
		},
	},
	{
		Name: "Meaning of life",
		CodeSteps: []step{
			{
				Line: `
				const (
					oneConst   = 1
					tenConst   = 10
					magicConst = 19
				)
				`,
				Result: "",
			},
			{
				Line:   "var outVar int",
				Result: "",
			},
			{
				Line: `
				type MyStruct struct {
					counter int
				}

				func (s *MyStruct) Add() {
					s.counter++
				}

				func (s *MyStruct) Get() int {
					return s.counter
				}
				`,
				Result: "",
			},
			{
				Line: `
				ms := &MyStruct{counter: 10}

				ms.Add()
				ms.Add()

				outVar = ms.Get() + oneConst + tenConst + magicConst

				println(outVar)
				`,
				Result: "42",
			},
		},
	},
}

func TestRepl(t *testing.T) {
	for _, fix := range fixtures {
		fix := fix
		t.Run(fix.Name, func(t *testing.T) {
			outbuf := new(bytes.Buffer)
			errbuf := new(bytes.Buffer)
			r := NewRepl(WithIO(os.Stdin, outbuf, errbuf))
			for _, cs := range fix.CodeSteps {
				r.RunStatements(cs.Line)
				errstr := stripTrailingNL(errbuf.String())
				require.Equal(t, cs.Error, errstr)
				outstr := stripTrailingNL(outbuf.String())
				require.Equal(t, cs.Result, outstr)

				// clear bufs.
				errbuf.Reset()
				outbuf.Reset()
			}
		})
	}
}

func stripTrailingNL(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s[:len(s)-1]
	} else {
		return s
	}
}

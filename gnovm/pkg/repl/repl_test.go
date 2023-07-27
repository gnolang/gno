package repl

import (
	"testing"

	"github.com/jaekwon/testify/require"
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
				Result: "import \"fmt\"\nimport \"os\"\n",
			},
		},
	},
	{
		Name: "Add wrong code",
		CodeSteps: []step{
			{
				Line:  "importasdasd",
				Error: "recovered from panic: test/test1.gno:7: name importasdasd not declared",
			},
			{
				Line: "var a := 1",
				// we cannot check the entire error because it is different depending on the used Go version.
				Error: "error parsing code:",
			},
		},
	},
	{
		Name: "Add function and use it",
		CodeSteps: []step{
			{
				Line:   "func sum(a,b int)int{return a+b}",
				Result: "func sum(a, b int) int\t{ return a + b }\n",
			},
			{
				Line:   "import \"fmt\"",
				Result: "import \"fmt\"\n",
			},
			{
				Line:   "fmt.Println(sum(1,1))",
				Result: "2\n",
			},
		},
	},
	{
		Name: "All declarations at once",
		CodeSteps: []step{
			{
				Line:   "import \"fmt\"\nfunc sum(a,b int)int{return a+b}",
				Result: "import \"fmt\"\nfunc sum(a, b int) int\t{ return a + b }\n",
			},
			{
				Line:   "fmt.Println(sum(1,1))",
				Result: "2\n",
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
				Result: "func fib(n int) int {\n\tif n < 2 {\n\t\treturn n\n\t}\n\treturn fib(n-2) + fib(n-1)\n}\n",
			},
			{
				Line:   "println(fib(24))",
				Result: "46368\n",
			},
		},
	},
}

func TestRepl(t *testing.T) {
	t.Parallel()
	for _, fix := range fixtures {
		t.Run(fix.Name, func(t *testing.T) {
			r := NewRepl()
			for _, cs := range fix.CodeSteps {
				out, err := r.Process(cs.Line)
				if cs.Error == "" {
					require.NoError(t, err)
				} else {
					require.Contains(t, err.Error(), cs.Error)
				}

				require.Equal(t, out, cs.Result)
			}
		})
	}
}

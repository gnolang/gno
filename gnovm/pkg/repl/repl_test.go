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
}

func TestRepl(t *testing.T) {
	for _, fix := range fixtures {
		t.Run(fix.Name, func(t *testing.T) {
			r := NewRepl()
			for _, cs := range fix.CodeSteps {
				out, err := r.Process(cs.Line)
				if cs.Error == "" {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}

				require.Equal(t, out, cs.Result)
			}

		})
	}
}

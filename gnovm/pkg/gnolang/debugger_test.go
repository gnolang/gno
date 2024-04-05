package gnolang_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/tests"
)

type dtest struct{ in, out string }

const debugTarget = "../../tests/integ/debugger/sample.gno"

type writeNopCloser struct{ io.Writer }

func (writeNopCloser) Close() error { return nil }

func eval(in, file string) (string, string) {
	out := bytes.NewBufferString("")
	err := bytes.NewBufferString("")
	stdin := bytes.NewBufferString(in)
	stdout := writeNopCloser{out}
	stderr := writeNopCloser{err}

	testStore := tests.TestStore(gnoenv.RootDir(), "", stdin, stdout, stderr, tests.ImportModeStdlibsPreferred)

	f := gnolang.MustReadFile(file)

	m := gnolang.NewMachineWithOptions(gnolang.MachineOptions{
		PkgPath: string(f.PkgName),
		Input:   stdin,
		Output:  stdout,
		Store:   testStore,
		Debug:   true,
	})

	defer m.Release()

	m.RunFiles(f)
	ex, _ := gnolang.ParseExpr("main()")
	m.Eval(ex)
	return out.String(), err.String()
}

func runDebugTest(t *testing.T, tests []dtest) {
	t.Helper()

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			out, err := eval(test.in, debugTarget)
			t.Log("in:", test.in, "out:", out, "err:", err)
			if !strings.Contains(out, test.out) {
				t.Errorf("result does not contain \"%s\", got \"%s\"", test.out, out)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	brk := "break 7\n"
	cont := brk + "continue\n"

	runDebugTest(t, []dtest{
		{in: "", out: "Welcome to the Gnovm debugger. Type 'help' for list of commands."},
		{in: "help\n", out: "The following commands are available"},
		{in: "h\n", out: "The following commands are available"},
		{in: "help h\n", out: "Print the help message."},
		{in: "list " + debugTarget + ":1\n", out: "1: // This is a sample target"},
		{in: brk, out: "Breakpoint 0 at main "},
		{in: cont, out: "=>    7: 	println(name, i)"},
		{in: cont + "stack\n", out: "2	in main.main"},
		{in: cont + "up\n", out: "=>   11: 	f(s, n)"},
		{in: cont + "print name\n", out: `("hello" string)`},
		{in: cont + "p i\n", out: `(3 int)`},
	})
}

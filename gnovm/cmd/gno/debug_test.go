package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type dtest struct{ in, out string }

const debugTarget = "../../tests/integ/debugger/sample.gno"

func runDebugTest(t *testing.T, tests []dtest) {
	args := []string{"run", "-debug", debugTarget}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			out := bytes.NewBufferString("")
			err := bytes.NewBufferString("")
			io := commands.NewTestIO()
			io.SetIn(bytes.NewBufferString(test.in))
			io.SetOut(commands.WriteNopCloser(out))
			io.SetErr(commands.WriteNopCloser(err))
			if err := newGnocliCmd(io).ParseAndRun(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			t.Log("out:", out)
			if !strings.Contains(out.String(), test.out) {
				t.Errorf("result does not contain \"%s\", got \"%s\"", test.out, out.String())
			}
		})
	}
}

func TestDebug(t *testing.T) {
	brk := "break " + debugTarget + ":7\n"
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

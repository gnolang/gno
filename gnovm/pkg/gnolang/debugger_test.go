package gnolang_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/tests"
)

type dtest struct{ in, out string }

const debugTarget = "../../tests/integ/debugger/sample.gno"

type writeNopCloser struct{ io.Writer }

func (writeNopCloser) Close() error { return nil }

func eval(debugAddr, in, file string) (string, string) {
	out := bytes.NewBufferString("")
	err := bytes.NewBufferString("")
	stdin := bytes.NewBufferString(in)
	stdout := writeNopCloser{out}
	stderr := writeNopCloser{err}

	testStore := tests.TestStore(gnoenv.RootDir(), "", stdin, stdout, stderr, tests.ImportModeStdlibsPreferred)

	f := gnolang.MustReadFile(file)

	m := gnolang.NewMachineWithOptions(gnolang.MachineOptions{
		PkgPath:   string(f.PkgName),
		Input:     stdin,
		Output:    stdout,
		Store:     testStore,
		Debug:     true,
		DebugAddr: debugAddr,
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
			out, err := eval("", test.in, debugTarget)
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
		{in: "l 40\n", out: "23: }"},
		{in: brk, out: "Breakpoint 0 at main "},
		{in: cont, out: "=>    7: 	println(name, i)"},
		{in: cont + "stack\n", out: "2	in main.main"},
		{in: cont + "up\n", out: "=>   11: 	f(s, n)"},
		{in: cont + "up\nup\ndown\n", out: "=>   11: 	f(s, n)"},
		{in: cont + "print name\n", out: `("hello" string)`},
		{in: cont + "p i\n", out: "(3 int)"},
		{in: cont + "bp\n", out: "Breakpoint 0 at main "},
		{in: "p 3\n", out: "(3 int)"},
		{in: "p 'a'\n", out: "(97 int32)"},
		{in: "p '界'\n", out: "(30028 int32)"},
		{in: "p \"xxxx\"\n", out: `("xxxx" string)`},
		{in: "si\n", out: "sample.gno:4"},
		{in: "s\ns\n", out: "=>   17: 	num := 5"},
		{in: "s\n\n", out: "=>   17: 	num := 5"},
		{in: "foo", out: "command not available: foo"},
		{in: "\n\n", out: "dbg> "},
		{in: "#\n", out: "dbg> "},
		{in: "p foo", out: "Command failed: could not find symbol value for foo"},
		{in: "b +7\nc\n", out: "=>   11:"},
		{in: brk + "clear 0\n", out: "dbg> "},
		{in: brk + "clear -1\n", out: "Command failed: invalid breakpoint id: -1"},
		{in: brk + "clear\n", out: "dbg> "},
		{in: "p\n", out: "Command failed: missing argument"},
		{in: "p 1+2\n", out: "Command failed: expression not supported"},
		{in: "p 1.2\n", out: "Command failed: invalid basic literal value: 1.2"},
		{in: "p 31212324222123123232123123123123123123123123123123\n", out: "value out of range"},
		{in: "p 3)\n", out: "Command failed:"},
	})
}

const debugAddress = "localhost:17358"

func TestRemoteDebug(t *testing.T) {
	var (
		conn  net.Conn
		err   error
		retry int
	)

	go eval(debugAddress, "", debugTarget)

	for retry = 100; retry > 0; retry-- {
		conn, err = net.Dial("tcp", debugAddress)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if retry == 0 {
		t.Error(err)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "d\n")
	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Error(err)
	}
	t.Log("resp:", resp)
}

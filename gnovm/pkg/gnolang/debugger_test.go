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

// TODO (Marc): move evalTest to gnovm/tests package and remove code duplicates
func evalTest(debugAddr, in, file string) (out, err, stacktrace string) {
	bout := bytes.NewBufferString("")
	berr := bytes.NewBufferString("")
	stdin := bytes.NewBufferString(in)
	stdout := writeNopCloser{bout}
	stderr := writeNopCloser{berr}
	debug := in != "" || debugAddr != ""
	mode := tests.ImportModeStdlibsPreferred
	if strings.HasSuffix(file, "_native.gno") {
		mode = tests.ImportModeNativePreferred
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Sprintf("%v", r)
		}
		out = strings.TrimSpace(out)
		err = strings.TrimSpace(strings.ReplaceAll(err, "../../tests/files/", "files/"))
	}()

	testStore := tests.TestStore(gnoenv.RootDir(), "../../tests/files", stdin, stdout, stderr, mode)

	f := gnolang.MustReadFile(file)

	m := gnolang.NewMachineWithOptions(gnolang.MachineOptions{
		PkgPath: string(f.PkgName),
		Input:   stdin,
		Output:  stdout,
		Store:   testStore,
		Context: tests.TestContext(string(f.PkgName), nil),
		Debug:   debug,
	})

	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case gnolang.UnhandledPanicError:
				stacktrace = m.ExceptionsStacktrace()
			default:
				stacktrace = m.Stacktrace().String()
			}
			stacktrace = strings.TrimSpace(strings.ReplaceAll(stacktrace, "../../tests/files/", "files/"))
			panic(r)
		}
	}()

	if debugAddr != "" {
		if e := m.Debugger.Serve(debugAddr); e != nil {
			err = e.Error()
			return
		}
	}

	m.RunFiles(f)
	ex, _ := gnolang.ParseExpr("main()")
	m.Eval(ex)
	out, err, stacktrace = bout.String(), berr.String(), m.ExceptionsStacktrace()
	return
}

func runDebugTest(t *testing.T, targetPath string, tests []dtest) {
	t.Helper()

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			out, err, _ := evalTest("", test.in, targetPath)
			t.Log("in:", test.in, "out:", out, "err:", err)
			if !strings.Contains(out, test.out) {
				t.Errorf("unexpected output\nwant\"%s\"\n  got \"%s\"", test.out, out)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	brk := "break 7\n"
	cont := brk + "continue\n"
	cont2 := "break 21\ncontinue\n"

	runDebugTest(t, debugTarget, []dtest{
		{in: "\n", out: "Welcome to the Gnovm debugger. Type 'help' for list of commands."},
		{in: "help\n", out: "The following commands are available"},
		{in: "h\n", out: "The following commands are available"},
		{in: "help b\n", out: "Set a breakpoint."},
		{in: "help zzz\n", out: "command not available"},
		{in: "list " + debugTarget + ":1\n", out: "1: // This is a sample target"},
		{in: "l 55\n", out: "46: }"},
		{in: "l xxx:0\n", out: "xxx: no such file or directory"},
		{in: "l :xxx\n", out: `"xxx": invalid syntax`},
		{in: brk, out: "Breakpoint 0 at main "},
		{in: "break :zzz\n", out: `"zzz": invalid syntax`},
		{in: "b +xxx\n", out: `"+xxx": invalid syntax`},
		{in: cont, out: "=>    7: 	println(name, i)"},
		{in: cont + "stack\n", out: "2	in main.main"},
		{in: cont + "up\n", out: "=>   11: 	f(s, n)"},
		{in: cont + "up\nup\ndown\n", out: "=>   11: 	f(s, n)"},
		{in: cont + "print name\n", out: `("hello" string)`},
		{in: cont + "p i\n", out: "(3 int)"},
		{in: cont + "up\np global\n", out: `("test" string)`},
		{in: cont + "bp\n", out: "Breakpoint 0 at main "},
		{in: "p 3\n", out: "(3 int)"},
		{in: "p 'a'\n", out: "(97 int32)"},
		{in: "p '界'\n", out: "(30028 int32)"},
		{in: "p \"xxxx\"\n", out: `("xxxx" string)`},
		{in: "si\n", out: "sample.gno:14"},
		{in: "s\ns\n", out: `=>   14: var global = "test"`},
		{in: "s\n\n", out: "=>   33: 	num := 5"},
		{in: "foo", out: "command not available: foo"},
		{in: "\n\n", out: "dbg> "},
		{in: "#\n", out: "dbg> "},
		{in: "p foo", out: "Command failed: could not find symbol value for foo"},
		{in: "b +7\nc\n", out: "=>   21: 	r := t.A[i]"},
		{in: brk + "clear 0\n", out: "dbg> "},
		{in: brk + "clear -1\n", out: "Command failed: invalid breakpoint id: -1"},
		{in: brk + "clear\n", out: "dbg> "},
		{in: "p\n", out: "Command failed: missing argument"},
		{in: "p 1+2\n", out: "Command failed: expression not supported"},
		{in: "p 1.2\n", out: "Command failed: invalid basic literal value: 1.2"},
		{in: "p 31212324222123123232123123123123123123123123123123\n", out: "value out of range"},
		{in: "p 3)\n", out: "Command failed:"},
		{in: "p (3)", out: "(3 int)"},
		{in: cont2 + "bt\n", out: "0	in main.(*main.T).get"},
		{in: cont2 + "p t.A[2]\n", out: "(3 int)"},
		{in: cont2 + "p t.A[k]\n", out: "could not find symbol value for k"},
		{in: cont2 + "p *t\n", out: "(struct{(slice[(1 int),(2 int),(3 int)] []int)} main.T)"},
		{in: cont2 + "p *i\n", out: "Not a pointer value: (1 int)"},
		{in: cont2 + "p *a\n", out: "could not find symbol value for a"},
		{in: cont2 + "p a[1]\n", out: "could not find symbol value for a"},
		{in: cont2 + "p t.B\n", out: "invalid selector: B"},
		{in: "down xxx", out: `"xxx": invalid syntax`},
		{in: "up xxx", out: `"xxx": invalid syntax`},
		{in: "b 37\nc\np b\n", out: "(3 int)"},
		{in: "b 27\nc\np b\n", out: `("!zero" string)`},
		{in: "b 22\nc\np t.A[3]\n", out: "Command failed: slice index out of bounds: 3 (len=3)"},
		{in: "b 43\nc\nc\nc\np i\ndetach\n", out: "(1 int)"},
	})

	runDebugTest(t, "../../tests/files/a1.gno", []dtest{
		{in: "l\n", out: "unknown source file"},
		{in: "b 5\n", out: "unknown source file"},
	})

	runDebugTest(t, "../../tests/integ/debugger/sample2.gno", []dtest{
		{in: "s\np tests\n", out: "(package(tests gno.land/p/demo/tests) package{})"},
		{in: "s\np tests.World\n", out: `("world" <untyped> string)`},
		{in: "s\np tests.xxx\n", out: "Command failed: invalid selector: xxx"},
	})
}

const debugAddress = "localhost:17358"

func TestRemoteDebug(t *testing.T) {
	var (
		conn  net.Conn
		err   error
		retry int
	)

	go evalTest(debugAddress, "", debugTarget)

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

func TestRemoteError(t *testing.T) {
	_, err, _ := evalTest(":xxx", "", debugTarget)
	t.Log("err:", err)
	if !strings.Contains(err, "tcp/xxx: unknown port") &&
		!strings.Contains(err, "tcp/xxx: nodename nor servname provided, or not known") {
		t.Error(err)
	}
}

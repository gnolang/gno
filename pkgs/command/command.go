package command

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
)

type Command struct {
	In     io.Reader
	InBuf  *bufio.Reader
	Out    io.WriteCloser
	OutBuf *bufio.Writer
	Err    io.WriteCloser
	ErrBuf *bufio.Writer
	Error  error
}

func NewStdCommand() *Command {
	cmd := new(Command)
	cmd.SetIn(os.Stdin) // needed for **** GetPassword().
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	return cmd
}

// An App does something with the *Command inputs and outputs.
// cmd: Command context.
// args: args to app.
// defaults: default options to app.
type App func(cmd *Command, args []string, defaults interface{}) error

// NOTE: defaults is first copied.
func (cmd *Command) Run(app App, args []string, defaults interface{}) error {
	args, flags := ParseArgs(args)
	ptr := amino.DeepCopyToPtr(defaults)
	err := applyFlags(ptr, flags)
	if err != nil {
		return err
	}
	opts := reflect.ValueOf(ptr).Elem().Interface()
	return app(cmd, args, opts)
}

func (cmd *Command) SetIn(in io.Reader) {
	cmd.In = in
	if inbuf, ok := cmd.In.(*bufio.Reader); ok {
		cmd.InBuf = inbuf
	} else {
		cmd.InBuf = bufio.NewReader(in)
	}
}

func (cmd *Command) SetOut(out io.WriteCloser) {
	cmd.Out = out
	cmd.OutBuf = bufio.NewWriter(cmd.Out)
}

func (cmd *Command) SetErr(err io.WriteCloser) {
	cmd.Err = err
	cmd.ErrBuf = bufio.NewWriter(cmd.Err)
}

//----------------------------------------
// NewMockCommand

// NewMockCommand returns a mock command for testing.
func NewMockCommand() *Command {
	mockIn := strings.NewReader("")
	mockOut := bytes.NewBufferString("")
	mockErr := bytes.NewBufferString("")
	cmd := new(Command)
	cmd.SetIn(mockIn)
	cmd.SetOut(WriteNopCloser(mockOut))
	cmd.SetErr(WriteNopCloser(mockErr))
	return cmd
}

type writeNopCloser struct {
	io.Writer
}

func (writeNopCloser) Close() error { return nil }

func WriteNopCloser(w io.Writer) io.WriteCloser {
	return writeNopCloser{w}

}

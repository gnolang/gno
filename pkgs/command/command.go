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
type App func(cmd *Command, args []string, opts interface{}) error

// defaults must be supplied for terminal apps only.
// NOTE: defaults is first copied, if provided.
func (cmd *Command) Run(app App, args []string, defaults interface{}) error {
	if defaults == nil {
		// for root/multi apps.
		return app(cmd, args, nil)
	} else {
		// for terminal apps.
		args, flags := ParseArgs(args)
		if help, ok := flags["help"]; ok && help == "y" {
			// print help.
			rt := reflect.TypeOf(defaults)
			cmd.printHelpFromDefaults(rt)
			return nil
		}
		// apply flags to defaults.
		ptr := amino.DeepCopyToPtr(defaults)
		err := applyFlags(ptr, flags)
		if err != nil {
			return err
		}
		opts := reflect.ValueOf(ptr).Elem().Interface()
		return app(cmd, args, opts)
	}
}

func (cmd *Command) printHelpFromDefaults(rt reflect.Type) {
	num := rt.NumField()

	// print anonymous embedded struct options
	for i := 0; i < num; i++ {
		rtf := rt.Field(i)
		if rtf.Anonymous {
			cmd.printHelpFromDefaults(rtf.Type)
			cmd.Println("")
		} else {
			continue
		}
	}

	// print remaining options
	cmd.Println("#", rt.Name(), "options")
	for i := 0; i < num; i++ {
		rtf := rt.Field(i)
		ffn := rtf.Tag.Get("flag")
		if rtf.Anonymous {
			continue
		} else if ffn == "" {
			// ignore fields with no flags field.
		} else {
			frt := rtf.Type
			help := rtf.Tag.Get("help")
			cmd.Println("-", ffn, "("+frt.String()+")", "-", help)
		}
	}

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

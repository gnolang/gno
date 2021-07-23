package command

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

type Command struct {
	Options interface{}
	Args    []string
	In      io.Reader
	InBuf   *bufio.Reader
	Out     io.WriteCloser
	OutBuf  *bufio.Writer
	Err     io.WriteCloser
	ErrBuf  *bufio.Writer
	Error   error
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

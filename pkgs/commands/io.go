package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// IO holds settable command
// input, output and error buffers
type IO struct {
	In    io.Reader
	inBuf *bufio.Reader

	Out    io.WriteCloser
	outBuf *bufio.Writer

	Err    io.WriteCloser
	errBuf *bufio.Writer
}

// DefaultIO returns a default command io
// that utilizes standard input / output / error
func DefaultIO() *IO {
	c := &IO{}

	c.SetIn(os.Stdin)
	c.SetOut(os.Stdout)
	c.SetErr(os.Stderr)

	return c
}

// SetIn sets the input reader for the command io
func (cmd *IO) SetIn(in io.Reader) {
	cmd.In = in
	if inbuf, ok := cmd.In.(*bufio.Reader); ok {
		cmd.inBuf = inbuf

		return
	}

	cmd.inBuf = bufio.NewReader(in)
}

// SetOut sets the output writer for the command io
func (cmd *IO) SetOut(out io.WriteCloser) {
	cmd.Out = out
	cmd.outBuf = bufio.NewWriter(cmd.Out)
}

// SetErr sets the error writer for the command io
func (cmd *IO) SetErr(err io.WriteCloser) {
	cmd.Err = err
	cmd.errBuf = bufio.NewWriter(cmd.Err)
}

// Println prints a line terminated by a newline
func (cmd *IO) Println(args ...interface{}) {
	if cmd.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintln(cmd.outBuf, args...)
	_ = cmd.outBuf.Flush()
}

// Printf prints a formatted string without trailing newline
func (cmd *IO) Printf(format string, args ...interface{}) {
	if cmd.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintf(cmd.outBuf, format, args...)
	_ = cmd.outBuf.Flush()
}

// Printfln prints a formatted string terminated by a newline
func (cmd *IO) Printfln(format string, args ...interface{}) {
	if cmd.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintf(cmd.outBuf, format+"\n", args...)
	_ = cmd.outBuf.Flush()
}

// ErrPrintln prints a line terminated by a newline to
// cmd.Err(Buf)
func (cmd *IO) ErrPrintln(args ...interface{}) {
	if cmd.errBuf == nil {
		return
	}

	_, _ = fmt.Fprintln(cmd.errBuf, args...)
	_ = cmd.errBuf.Flush()
}

// ErrPrintfln prints a formatted string terminated by a newline to cmd.Err(Buf)
func (cmd *IO) ErrPrintfln(format string, args ...interface{}) {
	if cmd.errBuf == nil {
		return
	}

	_, _ = fmt.Fprintf(cmd.errBuf, format+"\n", args...)
	_ = cmd.errBuf.Flush()
}

type writeNopCloser struct {
	io.Writer
}

func (writeNopCloser) Close() error { return nil }

func WriteNopCloser(w io.Writer) io.WriteCloser {
	return writeNopCloser{w}
}

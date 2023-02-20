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

// NewDefaultIO returns a default command io
// that utilizes standard input / output / error
func NewDefaultIO() *IO {
	c := &IO{}

	c.SetIn(os.Stdin)
	c.SetOut(os.Stdout)
	c.SetErr(os.Stderr)

	return c
}

// NewTestIO returns a test command io
// that only sets standard input (to avoid panics)
func NewTestIO() *IO {
	c := &IO{}
	c.SetIn(os.Stdin)

	return c
}

// SetIn sets the input reader for the command io
func (io *IO) SetIn(in io.Reader) {
	io.In = in
	if inbuf, ok := io.In.(*bufio.Reader); ok {
		io.inBuf = inbuf

		return
	}

	io.inBuf = bufio.NewReader(in)
}

// SetOut sets the output writer for the command io
func (io *IO) SetOut(out io.WriteCloser) {
	io.Out = out
	io.outBuf = bufio.NewWriter(io.Out)
}

// SetErr sets the error writer for the command io
func (io *IO) SetErr(err io.WriteCloser) {
	io.Err = err
	io.errBuf = bufio.NewWriter(io.Err)
}

// Println prints a line terminated by a newline
func (io *IO) Println(args ...interface{}) {
	if io.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintln(io.outBuf, args...)
	_ = io.outBuf.Flush()
}

// Printf prints a formatted string without trailing newline
func (io *IO) Printf(format string, args ...interface{}) {
	if io.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintf(io.outBuf, format, args...)
	_ = io.outBuf.Flush()
}

// Printfln prints a formatted string terminated by a newline
func (io *IO) Printfln(format string, args ...interface{}) {
	if io.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintf(io.outBuf, format+"\n", args...)
	_ = io.outBuf.Flush()
}

// ErrPrintln prints a line terminated by a newline to
// cmd.Err(Buf)
func (io *IO) ErrPrintln(args ...interface{}) {
	if io.errBuf == nil {
		return
	}

	_, _ = fmt.Fprintln(io.errBuf, args...)
	_ = io.errBuf.Flush()
}

// ErrPrintfln prints a formatted string terminated by a newline to cmd.Err(Buf)
func (io *IO) ErrPrintfln(format string, args ...interface{}) {
	if io.errBuf == nil {
		return
	}

	_, _ = fmt.Fprintf(io.errBuf, format+"\n", args...)
	_ = io.errBuf.Flush()
}

type writeNopCloser struct {
	io.Writer
}

func (writeNopCloser) Close() error { return nil }

func WriteNopCloser(w io.Writer) io.WriteCloser {
	return writeNopCloser{w}
}

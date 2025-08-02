package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// IO holds settable command
// input, output and error buffers
type IO interface {
	// getters
	In() io.Reader
	Out() io.WriteCloser
	Err() io.WriteCloser

	// setters and helpers
	SetIn(in io.Reader)
	SetOut(out io.WriteCloser)
	SetErr(err io.WriteCloser)
	Println(args ...any)
	Printf(format string, args ...any)
	Printfln(format string, args ...any)
	ErrPrintln(args ...any)
	ErrPrintfln(format string, args ...any)
	GetConfirmation(prompt string) (bool, error)
	GetPassword(prompt string, insecure bool) (string, error)
	GetString(prompt string) (string, error)
}

type IOImpl struct {
	in    io.Reader
	inBuf *bufio.Reader

	out    io.WriteCloser
	outBuf *bufio.Writer

	err    io.WriteCloser
	errBuf *bufio.Writer
}

// NewDefaultIO returns a default command io
// that utilizes standard input / output / error
func NewDefaultIO() IO {
	c := &IOImpl{}

	c.SetIn(os.Stdin)
	c.SetOut(os.Stdout)
	c.SetErr(os.Stderr)

	return c
}

// NewTestIO returns a test command io
// that only sets standard input (to avoid panics)
func NewTestIO() IO {
	c := &IOImpl{}
	c.SetIn(os.Stdin)

	return c
}

func (io *IOImpl) In() io.Reader       { return io.in }
func (io *IOImpl) Out() io.WriteCloser { return io.out }
func (io *IOImpl) Err() io.WriteCloser { return io.err }

// SetIn sets the input reader for the command io
func (io *IOImpl) SetIn(in io.Reader) {
	io.in = in
	if inbuf, ok := io.in.(*bufio.Reader); ok {
		io.inBuf = inbuf

		return
	}

	io.inBuf = bufio.NewReader(in)
}

// SetOut sets the output writer for the command io
func (io *IOImpl) SetOut(out io.WriteCloser) {
	io.out = out
	io.outBuf = bufio.NewWriter(io.out)
}

// SetErr sets the error writer for the command io
func (io *IOImpl) SetErr(err io.WriteCloser) {
	io.err = err
	io.errBuf = bufio.NewWriter(io.err)
}

// Println prints a line terminated by a newline
func (io *IOImpl) Println(args ...any) {
	if io.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintln(io.outBuf, args...)
	_ = io.outBuf.Flush()
}

// Printf prints a formatted string without trailing newline
func (io *IOImpl) Printf(format string, args ...any) {
	if io.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintf(io.outBuf, format, args...)
	_ = io.outBuf.Flush()
}

// Printfln prints a formatted string terminated by a newline
func (io *IOImpl) Printfln(format string, args ...any) {
	if io.outBuf == nil {
		return
	}

	_, _ = fmt.Fprintf(io.outBuf, format+"\n", args...)
	_ = io.outBuf.Flush()
}

// ErrPrintln prints a line terminated by a newline to
// cmd.Err(Buf)
func (io *IOImpl) ErrPrintln(args ...any) {
	if io.errBuf == nil {
		return
	}

	_, _ = fmt.Fprintln(io.errBuf, args...)
	_ = io.errBuf.Flush()
}

// ErrPrintfln prints a formatted string terminated by a newline to cmd.Err(Buf)
func (io *IOImpl) ErrPrintfln(format string, args ...any) {
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

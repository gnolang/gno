package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

type KeyPress byte

var (
	CRLF = []byte{'\r', '\n'}
	Null = []byte{0}
)

// key representation
const (
	KeyNone  KeyPress = 0      // None
	KeyCtrlC KeyPress = '\x03' // Ctrl+C
	KeyCtrlD KeyPress = '\x04' // Ctrl+D
	KeyCtrlE KeyPress = '\x05' // Ctrl+E
	KeyCtrlL KeyPress = '\x0c' // Ctrl+L
	KeyCtrlO KeyPress = '\x0f' // Ctrl+O
	KeyCtrlR KeyPress = '\x12' // Ctrl+R
	KeyCtrlT KeyPress = '\x14' // Ctrl+T
)

const (
	// ANSI escape codes
	ClearCurrentLine = "\033[2K"
	MoveCursorUp     = "\033[1A"
	MoveCursorDown   = "\033[1B"
)

func (k KeyPress) String() string {
	switch k {
	case KeyNone:
		return "Null"
	case KeyCtrlC:
		return "Ctrl+C"
	case KeyCtrlD:
		return "Ctrl+D"
	case KeyCtrlE:
		return "Ctrl+E"
	case KeyCtrlL:
		return "Ctrl+L"
	case KeyCtrlO:
		return "Ctrl+O"
	case KeyCtrlR:
		return "Ctrl+R"
	case KeyCtrlT:
		return "Ctrl+T"
		// For printable ASCII characters
	default:
		if k > 0x20 && k < 0x7e {
			return fmt.Sprintf("%c", k)
		}

		return fmt.Sprintf("Unknown (0x%02x)", byte(k))
	}
}

// rawTerminal wraps an io.Writer, converting \n to \r\n
type RawTerm struct {
	termMode     bool
	condTermMode *sync.Cond

	muTermMode sync.RWMutex

	fsin       *os.File
	reader     io.Reader
	taskWriter TaskWriter
}

func NewRawTerm() *RawTerm {
	return &RawTerm{
		condTermMode: sync.NewCond(&sync.Mutex{}),
		fsin:         os.Stdin,
		reader:       os.Stdin,
		taskWriter:   &rawTermWriter{os.Stdout},
	}
}

type restoreFunc func() error

func (rt *RawTerm) Init() (restoreFunc, error) {
	rt.muTermMode.Lock()
	defer rt.muTermMode.Unlock()

	fd := int(rt.fsin.Fd())
	oldstate, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("unable to init raw term: %w", err)
	}

	rt.reader = rt.fsin
	rt.taskWriter = &columnTermWriter{os.Stdout}
	return func() error {
		return term.Restore(fd, oldstate)
	}, nil
}

type TaskWriter interface {
	io.Writer
	WriteTask(task string, buf []byte) (n int, err error)
}

func (rt *RawTerm) enableTermMode() {
	rt.condTermMode.L.Lock()
	rt.termMode = true
	rt.fsin.Write([]byte{0}) // Release the key reader
	rt.condTermMode.L.Unlock()
}

func (rt *RawTerm) disableTermMode() {
	rt.condTermMode.L.Lock()
	rt.termMode = false
	rt.condTermMode.Broadcast()
	rt.condTermMode.L.Unlock()
}

func (rt *RawTerm) TermMode() <-chan string {
	rt.enableTermMode()

	rt.muTermMode.Lock()
	t := term.NewTerminal(rt.fsin, "> ")
	// Override output with terminal one
	rt.taskWriter = &rawTermWriter{t}
	rt.muTermMode.Unlock()

	// Create line reader chan
	rl := make(chan string)

	cleanup := func() {
		rt.muTermMode.Lock()
		// Jump one line
		fmt.Fprint(os.Stdout, "\r\n")

		// Set back reader/writer
		rt.taskWriter = &columnTermWriter{os.Stdout}

		// Signal that we are done
		close(rl)

		rt.muTermMode.Unlock()

		rt.disableTermMode()
	}

	go func() {
		defer cleanup()

		for {
			l, err := t.ReadLine()
			switch {
			case err == nil: // ok
			case !errors.Is(err, io.EOF):
				fmt.Fprintf(t, "error: %s\n", err.Error())
				fallthrough
			default:
				return
			}

			rl <- l
		}
	}()

	return rl
}

func (rt *RawTerm) Taskf(task string, format string, args ...interface{}) (n int, err error) {
	format = strings.TrimSpace(format)
	if len(args) > 0 {
		str := fmt.Sprintf(format, args...)
		return rt.taskWriter.WriteTask(task, []byte(str+"\n"))
	}

	return rt.taskWriter.WriteTask(task, []byte(format+"\n"))
}

func (rt *RawTerm) Task(task string) (n int, err error) {
	return rt.Taskf(task, "")
}

func (rt *RawTerm) Write(buf []byte) (n int, err error) {
	rt.muTermMode.RLock()
	defer rt.muTermMode.RUnlock()

	return rt.taskWriter.Write(buf)
}

func (rt *RawTerm) read(buf []byte) (n int, err error) {
	rt.muTermMode.RLock()
	defer rt.muTermMode.RUnlock()

	return rt.fsin.Read(buf)
}

func (rt *RawTerm) ReadKeyPress() (KeyPress, error) {
	for {
		rt.condTermMode.L.Lock()
		for rt.termMode {
			rt.condTermMode.Wait()
		}
		rt.condTermMode.L.Unlock()

		buf := make([]byte, 1)
		if _, err := rt.read(buf); err != nil {
			return KeyNone, err
		}

		return KeyPress(buf[0]), nil
	}
}

type columnTermWriter struct {
	writer io.Writer
}

func (r *columnTermWriter) Write(buf []byte) (n int, err error) {
	return r.WriteTask("", buf)
}

func (r *columnTermWriter) WriteTask(left string, buf []byte) (n int, err error) {
	for len(buf) > 0 {
		i := bytes.IndexByte(buf, '\n')
		todo := len(buf)
		if i >= 0 {
			todo = i
		}

		var nn int
		nn, err = r.writeColumnLine(left, buf[:todo])
		n += nn
		if err != nil {
			return n, err
		}
		buf = buf[todo:]

		if i >= 0 {
			if _, err = r.writer.Write(CRLF); err != nil {
				return n, err
			}
			n++
			buf = buf[1:]
		}
	}

	return
}

func (r *columnTermWriter) writeColumnLine(left string, line []byte) (n int, err error) {
	if left == "" {
		left = "."
	}

	// Write left column
	if n, err = fmt.Fprintf(r.writer, "%-15s | ", left); err != nil {
		return n, err
	}

	// Write left line
	var nn int
	nn, err = r.writer.Write(line)
	n += nn

	return
}

type rawTermWriter struct {
	writer io.Writer
}

func (r *rawTermWriter) Write(buf []byte) (n int, err error) {
	return r.writer.Write(buf)
}

func (r *rawTermWriter) WriteTask(task string, buf []byte) (n int, err error) {
	if task != "" {
		n, err = r.writer.Write([]byte(task + ": "))
	}

	var nn int
	nn, err = r.writer.Write(buf)
	n += nn
	return
}

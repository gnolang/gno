package dev

import (
	"bytes"
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
		taskWriter:   &rawTaskWriter{os.Stdout},
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
	rt.taskWriter = &columnTaskWriter{os.Stdout}
	return func() error {
		return term.Restore(fd, oldstate)
	}, nil
}

func (rt *RawTerm) Taskf(task string, format string, args ...interface{}) (n int, err error) {
	format = strings.TrimSpace(format)
	if len(args) > 0 {
		str := fmt.Sprintf(format, args...)
		return rt.taskWriter.WriteTask(task, []byte(str+"\n"))
	}

	return rt.taskWriter.WriteTask(task, []byte(format+"\n"))
}

func (rt *RawTerm) Write(buf []byte) (n int, err error) {
	rt.muTermMode.RLock()
	defer rt.muTermMode.RUnlock()

	return rt.taskWriter.Write(buf)
}

func (rt *RawTerm) WriteTask(name string, buf []byte) (n int, err error) {
	rt.muTermMode.RLock()
	defer rt.muTermMode.RUnlock()

	return rt.taskWriter.WriteTask(name, buf)
}

func (rt *RawTerm) NamespacedWriter(namepsace string) io.Writer {
	return &namespaceWriter{namepsace, rt}
}

func (rt *RawTerm) read(buf []byte) (n int, err error) {
	rt.muTermMode.RLock()
	defer rt.muTermMode.RUnlock()

	return rt.fsin.Read(buf)
}

func (rt *RawTerm) ReadKeyPress() (KeyPress, error) {
	buf := make([]byte, 1)
	if _, err := rt.read(buf); err != nil {
		return KeyNone, err
	}

	return KeyPress(buf[0]), nil
}

type namespaceWriter struct {
	namespace string
	writer    TaskWriter
}

func (r *namespaceWriter) Write(buf []byte) (n int, err error) {
	return r.writer.WriteTask(r.namespace, buf)
}

type TaskWriter interface {
	io.Writer
	WriteTask(task string, buf []byte) (n int, err error)
}

type columnTaskWriter struct {
	writer io.Writer
}

func (r *columnTaskWriter) Write(buf []byte) (n int, err error) {
	return r.WriteTask("", buf)
}

func (r *columnTaskWriter) WriteTask(left string, buf []byte) (n int, err error) {
	var nline int
	for nline = 0; len(buf) > 0; nline++ {
		i := bytes.IndexByte(buf, '\n')
		todo := len(buf)
		if i >= 0 {
			todo = i
		}

		var nn int
		switch {
		case nline == 0, left == "": // first line or left side is empty
			nn, err = r.writeColumnLine(left, buf[:todo])
		case i < 0 || i+1 == len(buf): // last line
			nn, err = r.writeColumnLine(" └─", buf[:todo])
		default: // middle lines
			nn, err = r.writeColumnLine(" │", buf[:todo])
		}

		n += nn
		if err != nil {
			return n, err
		}
		buf = buf[todo:]

		if i >= 0 { // always jump a line on the last line
			if _, err = r.writer.Write(CRLF); err != nil {
				return n, err
			}
			n++
			buf = buf[1:]
		}
	}

	return
}

func (r *columnTaskWriter) writeColumnLine(left string, line []byte) (n int, err error) {
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

type rawTaskWriter struct {
	writer io.Writer
}

func (r *rawTaskWriter) Write(buf []byte) (n int, err error) {
	return r.writer.Write(buf)
}

func (r *rawTaskWriter) WriteTask(task string, buf []byte) (n int, err error) {
	if task != "" {
		n, err = r.writer.Write([]byte(task + ": "))
	}

	var nn int
	nn, err = r.writer.Write(buf)
	n += nn
	return
}

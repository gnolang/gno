package rawterm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

var CRLF = []byte{'\r', '\n'}

// rawTerminal wraps an io.Writer, converting \n to \r\n
type RawTerm struct {
	syncWriter sync.Mutex

	fsin   *os.File
	reader io.Reader
	writer io.Writer
}

func NewRawTerm() *RawTerm {
	return &RawTerm{
		fsin:   os.Stdin,
		reader: os.Stdin,
		writer: os.Stdout,
	}
}

func (rt *RawTerm) Init() (restore func() error, err error) {
	fd := int(rt.fsin.Fd())
	oldstate, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("unable to init raw term: %w", err)
	}

	rt.reader = rt.fsin
	return func() error {
		return term.Restore(fd, oldstate)
	}, nil
}

func (rt *RawTerm) Write(buf []byte) (n int, err error) {
	rt.syncWriter.Lock()
	defer rt.syncWriter.Unlock()

	return writeWithCRLF(rt.writer, buf)
}

func (rt *RawTerm) read(buf []byte) (n int, err error) {
	return rt.fsin.Read(buf)
}

func (rt *RawTerm) ReadKeyPress() (KeyPress, error) {
	buf := make([]byte, 1)
	if _, err := rt.read(buf); err != nil {
		return KeyNone, err
	}

	return KeyPress(buf[0]), nil
}

// writeWithCRLF writes buf to w but replaces all occurrences of \n with \r\n.
func writeWithCRLF(w io.Writer, buf []byte) (n int, err error) {
	for len(buf) > 0 {
		i := bytes.IndexByte(buf, '\n')
		todo := len(buf)
		if i >= 0 {
			todo = i
		}

		var nn int
		nn, err = w.Write(buf[:todo])
		n += nn
		if err != nil {
			return n, err
		}
		buf = buf[todo:]

		if i >= 0 {
			if _, err = w.Write(CRLF); err != nil {
				return n, err
			}
			n++
			buf = buf[1:]
		}
	}

	return n, nil
}

// type namespaceWriter struct {
// 	namespace string
// 	writer    TaskWriter
// }

// func (r *namespaceWriter) Write(buf []byte) (n int, err error) {
// 	return r.writer.WriteTask(r.namespace, buf)
// }

// type TaskWriter interface {
// 	io.Writer
// 	WriteTask(task string, buf []byte) (n int, err error)
// }

// type columnTaskWriter struct {
// 	writer io.Writer
// }

// func (r *columnTaskWriter) Write(buf []byte) (n int, err error) {
// 	return r.WriteTask("", buf)
// }

// func (r *columnTaskWriter) WriteTask(left string, buf []byte) (n int, err error) {
// 	var nline int
// 	for nline = 0; len(buf) > 0; nline++ {
// 		i := bytes.IndexByte(buf, '\n')
// 		todo := len(buf)
// 		if i >= 0 {
// 			todo = i
// 		}

// 		var nn int
// 		switch {
// 		case nline == 0, left == "": // first line or left side is empty
// 			nn, err = r.writeColumnLine(left, buf[:todo])
// 		case i < 0 || i+1 == len(buf): // last line
// 			nn, err = r.writeColumnLine(" └─", buf[:todo])
// 		default: // middle lines
// 			nn, err = r.writeColumnLine(" │", buf[:todo])
// 		}

// 		n += nn
// 		if err != nil {
// 			return n, err
// 		}
// 		buf = buf[todo:]

// 		if i >= 0 { // always jump a line on the last line
// 			if _, err = r.writer.Write(CRLF); err != nil {
// 				return n, err
// 			}
// 			n++
// 			buf = buf[1:]
// 		}
// 	}

// 	return
// }

// func (r *columnTaskWriter) writeColumnLine(left string, line []byte) (n int, err error) {
// 	// Write left column
// 	if n, err = fmt.Fprintf(r.writer, "%-15s | ", left); err != nil {
// 		return n, err
// 	}

// 	// Write left line
// 	var nn int
// 	nn, err = r.writer.Write(line)
// 	n += nn

// 	return
// }

// type rawTaskWriter struct {
// 	writer io.Writer
// }

// func (r *rawTaskWriter) Write(buf []byte) (n int, err error) {
// 	return r.writer.Write(buf)
// }

// func (r *rawTaskWriter) WriteTask(task string, buf []byte) (n int, err error) {
// 	if task != "" {
// 		n, err = r.writer.Write([]byte(task + ": "))
// 	}

// 	var nn int
// 	nn, err = r.writer.Write(buf)
// 	n += nn
// 	return
// }

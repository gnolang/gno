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

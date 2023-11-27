package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/term"
)

type KeyPress byte

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
	muTermMode sync.RWMutex
	fsin       *os.File

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

type restoreFunc func() error

func (rt *RawTerm) Init() (restoreFunc, error) {
	rt.muTermMode.Lock()
	defer rt.muTermMode.Unlock()

	fd := int(rt.fsin.Fd())
	oldstate, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("unable to init raw term: %w", err)
	}

	rt.reader = os.Stdin
	rt.writer = &rawTermWriter{os.Stdout}
	return func() error {
		return term.Restore(fd, oldstate)
	}, nil
}

func (rt *RawTerm) bindInputOutput() {

	rt.muTermMode.Unlock()
}

func (rt *RawTerm) TermMode() <-chan string {
	rt.muTermMode.Lock()
	t := term.NewTerminal(rt.fsin, "> ")

	// create a blocking reader to prevent external read
	noopreader := newBlockingReader()

	// override input/output with terminal ones
	rt.writer = t
	rt.reader = noopreader
	rt.muTermMode.Unlock()

	// create line reader chan
	rl := make(chan string)

	cleanup := func() {
		rt.muTermMode.Lock()

		// set back reader/writer
		rt.reader = os.Stdin
		rt.writer = &rawTermWriter{os.Stdout}
		// cleanup output
		fmt.Fprintf(rt.writer, "\r\n")
		// unlock pending reader
		noopreader.Close()
		// signal that we are done
		close(rl)

		rt.muTermMode.Unlock()
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

// Write implements the io.Writer interface for rawModeWriter.
// It converts each \n in the input to \r\n.
func (rt *RawTerm) Write(buf []byte) (n int, err error) {
	rt.muTermMode.RLock()
	defer rt.muTermMode.RUnlock()

	return rt.writer.Write(buf)
}

func (rt *RawTerm) Read(buf []byte) (n int, err error) {
	rt.muTermMode.RLock()
	defer rt.muTermMode.RUnlock()

	return rt.reader.Read(buf)
}

func (rt *RawTerm) ReadKeyPress() (KeyPress, error) {
	buf := make([]byte, 1)
	if _, err := rt.Read(buf); err != nil {
		return KeyNone, err
	}

	return KeyPress(buf[0]), nil
}

func listenForKeyPress(io commands.IO, rt *RawTerm) <-chan KeyPress {
	cc := make(chan KeyPress)
	go func() {
		defer close(cc)
		key, err := rt.ReadKeyPress()
		if err != nil {
			io.ErrPrintfln("unable to read keypress: %s", err.Error())
			return
		}

		cc <- key
	}()

	return cc
}

type rawTermWriter struct {
	writer io.Writer
}

func (r *rawTermWriter) Write(p []byte) (n int, err error) {
	modified := bytes.ReplaceAll(p, []byte{'\n'}, []byte{'\r', '\n'})
	return r.writer.Write(modified)
}

// NoopReader defines a reader that does nothing
type noopReader struct {
	once sync.Once
	cc   chan struct{}
}

func newBlockingReader() *noopReader {
	return &noopReader{
		cc: make(chan struct{}),
	}
}

func (n *noopReader) Close() error {
	n.once.Do(func() { close(n.cc) })
	return nil
}

func (n *noopReader) Read(p []byte) (int, error) {
	<-n.cc
	return 0, io.EOF
}

// type atomicBool struct {
// 	*sync.Cond
// 	flag int32
// }

// func NewAtomicBool() *atomicBool {
// 	return &atomicBool{
// 		Cond: sync.NewCond(&sync.Mutex{}),
// 		flag: 0,
// 	}
// }

// func (ab *atomicBool) SetTrue() *atomicBool {
// 	if atomic.CompareAndSwapInt32(&ab.flag, 0, 1) {
// 		// Only notify if there was a change
// 		ab.Broadcast()
// 	}

// 	return ab
// }

// func (ab *atomicBool) SetFalse() *atomicBool {
// 	if atomic.CompareAndSwapInt32(&ab.flag, 1, 0) {
// 		// Only notify if there was a change
// 		ab.Broadcast()
// 	}
// 	return ab
// }

// func (ab *atomicBool) WaitUntilFalse() {
// 	ab.L.Lock()
// 	for atomic.LoadInt32(&ab.flag) != 0 {
// 		ab.Wait()
// 	}
// 	ab.L.Unlock()
// }

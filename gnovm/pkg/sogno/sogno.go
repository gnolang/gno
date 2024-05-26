// Package sogno implements a runtime for Gno programs in Go, so that transpiled
// code can be called as a binary.
package sogno

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"reflect"

	"github.com/tidwall/btree"
)

// Context contains the runtime information of the running Gno program.
type Context struct {
	// Import path of this realm.
	Path string

	// Map of the global variables in this realm.
	// Each should be associated with a pointer to the value.
	State *btree.Map[string, interface{}]

	input  io.Reader
	output io.Writer
}

// Main should be called from the main function of each realm.
//
// The main program of a Sogno binary is a text-based, newline-separated
// interface to the realm, for ease of use by humans and machines.
// The supported commands are the following:
// (TODO, this is a lie, see the code for what is true)
//
//	load <identifier> <value>
//	read <identifier>
//		read the value of the state variable with the given identifier, encoded
//		as binary then hex.
//	call <function_name> [<args...>]
//	debug context
//		show the current value of the context:
//		- import path
//		- state variables
func (c *Context) Main() {
	const (
		bufSize = 4 << 10
	)

	// buffer for reading from stdin
	buf := make([]byte, bufSize)
	if c.input == nil && c.output == nil {
		c.input, c.output = os.Stdin, os.Stdout
	}

	read, err := c.input.Read(buf)
	if err != nil {
		c.readError(err)
	}
	buf = buf[:read]

	for {
		nl := bytes.IndexByte(buf, '\n')
		if nl < 0 {
			// Newline not found; grow buf if necessary and read
			// until it is found.
			if len(buf) == cap(buf) {
				buf = append(make([]byte, 0, cap(buf)*2), buf...)
			}

			read, err = c.input.Read(buf[len(buf):cap(buf)])
			if err != nil {
				c.readError(err)
			}
			buf = buf[:len(buf)+read]
			continue
		}

		// parse line and command
		line := buf[:nl]
		sp := bytes.IndexByte(line, ' ')
		if sp < 0 {
			sp = nl
		}
		command := string(buf[:sp])

		// main switch
		switch command {
		case "debug":
			c.commandDebug()
		case "read":
			c.commandRead(line[sp+1:])
		default:
			_, err = c.output.Write([]byte("bad command: " + command + "\n"))
		}

		// remove line from buf
		delta := len(buf) - (nl + 1)
		copy(buf[:delta], buf[nl+1:])
		buf = buf[:delta]
	}
}

func (c *Context) commandDebug() {
	var buf bytes.Buffer
	buf.WriteString("import path: " + c.Path + "\n")
	buf.WriteString("state variables:\n")

	keys := c.State.Keys()
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteByte('\n')
	}
	_, err := c.output.Write(buf.Bytes())
	if err != nil {
		c.writeError(err)
	}
}

func (c *Context) commandRead(arg []byte) {
	argStr := string(arg)
	val, set := c.State.Get(argStr)
	if !set {
		_, err := c.output.Write([]byte("var " + argStr + " undefined\n"))
		if err != nil {
			c.writeError(err)
		}
		return
	}
	v := reflect.ValueOf(val).Elem()
	t := v.Type()
	_, err := c.output.Write([]byte("var " + argStr + " " + t.String() + "\n"))
	if err != nil {
		c.writeError(err)
	}
	var buf bytes.Buffer
	enc := hex.NewEncoder(&buf)
	err = binary.Write(enc, binary.LittleEndian, v.Interface())
	if err != nil {
		_, err := c.output.Write([]byte("encoding error: " + err.Error() + "\n"))
		if err != nil {
			c.writeError(err)
		}
		return
	}
	buf.WriteByte('\n')
	_, err = c.output.Write(buf.Bytes())
	if err != nil {
		c.writeError(err)
	}
}

func (c *Context) readError(err error) {
	if errors.Is(err, io.EOF) {
		os.Exit(0)
	}
	_, _ = os.Stderr.WriteString("stdin read error: " + err.Error() + "\n")
	os.Exit(1)
}

func (c *Context) writeError(err error) {
	_, _ = os.Stderr.WriteString("stdout write error: " + err.Error() + "\n")
	os.Exit(1)
}

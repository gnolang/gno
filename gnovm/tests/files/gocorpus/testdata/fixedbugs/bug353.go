// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 2089 - internal compiler error

package main

import (
	"io"
	"os"
)

func echo(fd io.ReadWriterCloser) { // ERROR "undefined.*io.ReadWriterCloser"
	var buf [1024]byte
	for {
		n, err := fd.Read(buf)
		if err != nil {
			break
		}
		fd.Write(buf[0:n])
	}
}

func main() {
	fd, _ := os.Open("a.txt")
	echo(fd)
}

// GnoError:
// line 16: name ReadWriterCloser not declared
// line 18: expected declaration, found 'for'
// line 19: expected declaration, found n
// line 20: expected declaration, found 'if'
// line 21: expected declaration, found 'break'
// line 22: expected declaration, found '}'
// line 23: expected declaration, found fd
// line 24: expected declaration, found '}'
// line 25: expected declaration, found '}'
// line 28: name Open not declared

// GoTypeCheckError:
// line 16: undefined: io.ReadWriterCloser

// GnoOverStrictError:
// line 18: expected declaration, found 'for'
// line 19: expected declaration, found n
// line 20: expected declaration, found 'if'
// line 21: expected declaration, found 'break'
// line 22: expected declaration, found '}'
// line 23: expected declaration, found fd
// line 24: expected declaration, found '}'
// line 25: expected declaration, found '}'
// line 28: name Open not declared

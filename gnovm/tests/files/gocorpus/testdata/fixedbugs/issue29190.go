// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"strings"
)

type T struct{}

const maxInt = int(^uint(0) >> 1)

func main() {
	s := make([]T, maxInt)
	shouldPanic("len out of range", func() { s = append(s, T{}) })
	var oneElem = make([]T, 1)
	shouldPanic("len out of range", func() { s = append(s, oneElem...) })
}

func shouldPanic(str string, f func()) {
	defer func() {
		err := recover()
		if err == nil {
			panic("did not panic")
		}
		s := err.(error).Error()
		if !strings.Contains(s, str) {
			panic("got panic " + s + ", want " + str)
		}
	}()

	f()
}

// GnoOutput:

// GnoError:
// multiplication overflow

// GoOutput:

// KnownDivergence:
// gc's make limit is byte-based and zero-sized elements cost 0 bytes, so
// make([]T, maxInt) is legal there. GnoVM boxes every element as a
// TypedValue (NewListArray: make([]TypedValue, n)), so zero-sized types
// aren't free and the per-element cap IS its allocatable-size limit —
// spec-permitted. Since #5723 the panic is recoverable with the Go-style
// "makeslice: len out of range" message. KnownDivergence candidate.

// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"runtime"
)

type HeapObj [8]int64

type StkObj struct {
	h *HeapObj
}

var n int
var c int = -1

func gc() {
	// encourage heap object to be collected, and have its finalizer run.
	runtime.GC()
	runtime.GC()
	runtime.GC()
	n++
}

func main() {
	f()
	gc() // prior to stack objects, heap object is not collected until here
	if c < 0 {
		panic("heap object never collected")
	}
	if c != 1 {
		panic(fmt.Sprintf("expected collection at phase 1, got phase %d", c))
	}
}

func f() {
	var s StkObj
	s.h = new(HeapObj)
	runtime.SetFinalizer(s.h, func(h *HeapObj) {
		// Remember at what phase the heap object was collected.
		c = n
	})
	g(&s)
	gc()
}

func g(s *StkObj) {
	gc() // heap object is still live here
	runtime.KeepAlive(s)
	gc() // heap object should be collected here
}

// TypeCheckError:
// main/stackobj.go:45:10: undefined: runtime.SetFinalizer; main/stackobj.go:55:10: undefined: runtime.KeepAlive

// GnoOutput:

// GnoError:
// main/stackobj.go:45:2-22: name SetFinalizer not declared

// GoOutput:

// KnownIssue:
// TODO: explain the Gno bug (Gno errors where Go runs clean)

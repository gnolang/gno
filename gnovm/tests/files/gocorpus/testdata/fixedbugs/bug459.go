// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 3890: missing detection of init cycle involving
// method calls in function bodies.

package flag

var commandLine = NewFlagSet() // ERROR "initialization cycle|depends upon itself"

type FlagSet struct {
}

func (f *FlagSet) failf(format string, a ...interface{}) {
	f.usage()
}

func (f *FlagSet) usage() {
	if f == commandLine {
		panic(3)
	}
}

func NewFlagSet() *FlagSet {
	f := &FlagSet{}
	f.setErrorHandling(true)
	return f
}

func (f *FlagSet) setErrorHandling(b bool) {
	f.failf("DIE")
}

// GoTypeCheckError:
// line 12
// line 17
// line 21
// line 27
// line 33

// GoTypesOnly:
// line 12: gc expects: initialization cycle|depends upon itself

// Verdict: Clean errorcheck

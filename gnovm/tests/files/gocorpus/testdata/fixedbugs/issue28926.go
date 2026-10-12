// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Stringer interface {
	String() string
}

func main() {
	var e interface{}
	switch e := e.(type) {
	case G: // ERROR "undefined: G|undefined type .*G"
		e.M() // ok: this error should be ignored because the case failed its typecheck
	case E: // ERROR "undefined: E|undefined type .*E"
		e.D() // ok: this error should be ignored because the case failed its typecheck
	case Stringer:
		// ok: this error should not be ignored to prove that passing legs aren't left out
		_ = e.(T) // ERROR "undefined: T|undefined type .*T"
	}
}

// GnoError:
// line 16: name G not declared

// GoTypeCheckError:
// line 16
// line 18
// line 22

// GoTypesOnly:
// line 18: gc expects: undefined: E|undefined type .*E
// line 22: gc expects: undefined: T|undefined type .*T

// Verdict: Clean errorcheck

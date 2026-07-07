// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that we get the correct (T vs &T) literal specification
// in the error message.

package p

type S struct {
	f T
}

type P struct {
	f *T
}

type T struct{}

var _ = S{
	f: &T{}, // ERROR "cannot use &T{}|incompatible type"
}

var _ = P{
	f: T{}, // ERROR "cannot use T{}|incompatible type"
}

// GnoError:
// line 22: 2: cannot use *gno.land/p/filetest/p.T as struct{}
// line 23: expected declaration, found f
// line 24: expected declaration, found '}'
// line 26: 2: cannot use struct{} as *gno.land/p/filetest/p.T
// line 27: expected declaration, found f
// line 28: expected declaration, found '}'

// GoTypeCheckError:
// line 23: cannot use &T{} (value of type *T) as T value in struct literal
// line 27: cannot use T{} (value of struct type T) as *T value in struct literal

// GnoOverStrictError:
// line 22: 2: cannot use *gno.land/p/filetest/p.T as struct{}
// line 24: expected declaration, found '}'
// line 26: 2: cannot use struct{} as *gno.land/p/filetest/p.T
// line 28: expected declaration, found '}'

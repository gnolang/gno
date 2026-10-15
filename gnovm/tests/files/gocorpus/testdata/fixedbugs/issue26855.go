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
// line 26: 2: cannot use struct{} as *gno.land/p/filetest/p.T

// GnoWordingDiffers:
// line 23: gc expects: cannot use &T{}|incompatible type; gno: 2: cannot use *gno.land/p/filetest/p.T as struct{}
// line 27: gc expects: cannot use T{}|incompatible type; gno: 2: cannot use struct{} as *gno.land/p/filetest/p.T

// Verdict: Clean errorcheck

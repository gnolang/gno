// errorcheck

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type I interface {
	F()
}

type T struct {
}

const _ = I((*T)(nil)) // ERROR "is not constant"

func (*T) F() {
}

// GnoError:
// line 16: (const (nil *gno.land/p/filetest/p.T)) (variable of type gno.land/p/filetest/p.I) is not constant

// GoTypeCheckError:
// line 16: I((*T)(nil)) (value of interface type I) is not constant

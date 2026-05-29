// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() {
	1 = 2 // ERROR "cannot assign to 1|invalid left hand side"
}

// GnoError:
// line 10: cannot assign to (const (1 <untyped> bigint))

// GoTypeCheckError:
// line 10: cannot assign to 1 (neither addressable nor a map index expression)

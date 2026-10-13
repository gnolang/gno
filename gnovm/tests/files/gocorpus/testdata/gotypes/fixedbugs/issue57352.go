// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type A interface {
	a()
}

type AB interface {
	A
	b()
}

type AAB struct {
	A
	AB
}

var _ AB = AAB /* ERROR "ambiguous selector AAB.a" */ {}

// GnoError:
// line 21: gno.land/p/filetest/p.AAB does not implement gno.land/p/filetest/p.AB (missing method a)

// GoTypeCheckError:
// line 21

// Clean: errorcheck

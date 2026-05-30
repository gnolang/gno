// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// cmd/compile erroneously rejected conversions of constant values
// between int/float and complex types.

package p

const (
	_ = int(complex64(int(0)))
	_ = float64(complex128(float64(0)))

	_ = int8(complex128(1000)) // ERROR "overflow|cannot convert"
)

// GoTypeCheckError:
// line 16: cannot convert complex128(1000) (constant 1000 of type complex128) to type int8

// KnownIssue:
// line 13: name complex64 not defined in fileset with files [issue38117.go]

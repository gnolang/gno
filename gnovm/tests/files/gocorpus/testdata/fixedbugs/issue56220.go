// errorcheck

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() int {
	return int(1 - .0000001) // ERROR "cannot convert 1 - \.0000001 \(untyped float constant 0\.9999999\) to type int"
}

func g() int64 {
	return int64((float64(0.03) - float64(0.02)) * 1_000_000) // ERROR "cannot convert \(float64\(0\.03\) - float64\(0\.02\)\) \* 1_000_000 \(constant 9999\.999999999998 of type float64\) to type int64"
}

// GnoError:
// line 10: cannot convert (const (0.9999999 <untyped> bigdec)) to integer type

// GoTypeCheckError:
// line 10: cannot convert 1 - .0000001 (untyped float constant 0.9999999) to type int
// line 14: cannot convert (float64(0.03) - float64(0.02)) * 1_000_000 (constant 9999.999999999998 of type float64) to type int64

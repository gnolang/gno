// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	var i33 int64;
	if i33 == (1<<64) -1 {  // ERROR "overflow"
	}
}

// GnoError:
// line 11: bigint overflows target kind
// line 13: expected declaration, found '}'

// GoTypeCheckError:
// line 11: (1 << 64) - 1 (untyped int constant 18446744073709551615) overflows int64

// GnoOverStrictError:
// line 13: expected declaration, found '}'

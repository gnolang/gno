// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
func f() string {
	return 0	// ERROR "conversion|type"
}

// GnoError:
// line 9: cannot use untyped Bigint as StringKind

// GoTypeCheckError:
// line 9: cannot use 0 (untyped int constant) as string value in return statement

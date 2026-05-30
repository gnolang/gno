// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
func main() {
	var v interface{} = 0;
	switch v.(type) {
	case int:
		fallthrough;		// ERROR "fallthrough"
	default:
		panic("fell through");
	}
}

// GnoError:
// line 12: cannot fallthrough in type switch

// GoTypeCheckError:
// line 12: cannot fallthrough in type switch

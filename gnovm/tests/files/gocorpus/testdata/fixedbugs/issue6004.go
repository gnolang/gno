// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	_ = nil                  // ERROR "use of untyped nil"
	_, _ = nil, 1            // ERROR "use of untyped nil"
	_, _ = 1, nil            // ERROR "use of untyped nil"
	_ = append(nil, 1, 2, 3) // ERROR "untyped nil|nil"
}

// GnoError:
// line 10: use of untyped nil in assignment
// line 11: use of untyped nil in assignment
// line 12: use of untyped nil in assignment
// line 13: first argument to append must be a typed slice; have untyped nil

// GoTypeCheckError:
// line 10
// line 11
// line 12
// line 13

// Verdict: Clean errorcheck

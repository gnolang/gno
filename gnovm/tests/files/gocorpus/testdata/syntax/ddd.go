// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f() {
	g(f..3) // ERROR "unexpected literal \.3, expected name or \("
}

// GnoError:
// line 10: expected selector or type assertion, found .3

// GoTypeCheckError:
// line 10: expected selector or type assertion, found .3

// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Used to die dividing by zero; issue 879.

package main

var mult [3][...]byte = [3][5]byte{}	// ERROR "\.\.\."

// GnoError:
// line 11: cannot use [3][5]uint8 as [3][0]uint8

// GoTypeCheckError:
// line 11: invalid use of [...] array (outside a composite literal)

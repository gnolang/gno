// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var x⊛y int	// ERROR "invalid character .* in identifier"

// GnoError:
// line 9: expected type, found 'ILLEGAL'

// GnoWordingDiffers:
// line 9: gc expects: invalid character .* in identifier; gno: expected type, found 'ILLEGAL'

// Verdict: Clean errorcheck

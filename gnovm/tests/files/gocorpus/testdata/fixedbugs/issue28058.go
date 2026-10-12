// errorcheck

// Copyright 2018 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 14988: declaring a map with an invalid key type should not cause a
//              fatal panic.

package main

var x map[func()]int // ERROR "invalid map key type"
var X map[func()]int // ERROR "invalid map key type"

// GoTypeCheckError:
// line 12
// line 13

// GoTypesOnly:
// line 12: gc expects: invalid map key type
// line 13: gc expects: invalid map key type

// Verdict: Clean errorcheck

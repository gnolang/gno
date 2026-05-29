// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// identifiers beginning with non-ASCII digits were incorrectly accepted.
// issue 11359.

package p
var ۶ = 0 // ERROR "identifier cannot begin with digit"

// GnoError:
// line 11: illegal character U+06F6 '۶'

// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test an internal compiler error on ? symbol in declaration
// following an empty import.

package a
import""  // ERROR "import path is empty|invalid import path \(empty string\)"

// GnoError:
// line 11: invalid zero package path in testStore().pkgGetter:
// --- preprocess stack ---
// stack 1: file gno.land/p/filetest/a/issue11610a.go:11:1-14:15
// ------------------------

// Verdict: Clean errorcheck

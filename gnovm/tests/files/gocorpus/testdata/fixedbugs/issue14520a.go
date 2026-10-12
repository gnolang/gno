// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f

import /* // ERROR "import path" */ `
bogus`

// GnoError:
// line 9: 7: unknown import path 
// bogus:
// --- preprocess stack ---
// stack 1: file gno.land/p/filetest/f/issue14520a.go:8:1-13:15
// ------------------------

// GoTypeCheckError:
// line 9

// Verdict: Clean errorcheck

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T struct{}
type I interface{ M() }
var _ I = T /* ERROR "missing method M" */ {} // must not crash
func (T) m() {}

// GnoError:
// line 9: gno.land/p/filetest/p.T does not implement gno.land/p/filetest/p.I (missing method M)

// Verdict: Clean errorcheck

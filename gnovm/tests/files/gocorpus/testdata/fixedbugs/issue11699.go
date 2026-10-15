// compile

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 11699; used to fail with duplicate _.args_stackmap symbols.

package p

func _()
func _()

// Verdict: Unsupported gc-specific: gc's bodiless function declaration (an assembly body) has no Gno equivalent

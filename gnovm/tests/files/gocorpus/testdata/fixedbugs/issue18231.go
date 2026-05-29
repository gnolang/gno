// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that error message for composite literals with
// missing type is at the right place.

package p

type T struct {
	f map[string]string
}

var _ = T{
	f: {                // ERROR "missing type in composite literal|may only omit types within"
		"a": "b",
	},
}

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 16: 2: types cannot be elided in composite literals for struct types

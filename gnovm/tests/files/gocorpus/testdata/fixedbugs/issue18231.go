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

// GnoError:
// line 16: 2: types cannot be elided in composite literals for struct types

// GoTypeCheckError:
// line 17: missing type in composite literal

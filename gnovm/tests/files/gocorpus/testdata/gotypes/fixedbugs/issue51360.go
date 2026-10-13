// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() {
	len.Println /* ERROR "invalid use of len (built-in) in selector expression" */
	len.Println /* ERROR "invalid use of len (built-in) in selector expression" */ ()
	_ = len.Println /* ERROR "invalid use of len (built-in) in selector expression" */
	_ = len /* ERROR "cannot index len" */ [0]
	_ = *len /* ERROR "cannot indirect len" */
}

// GnoError:
// line 8: use of builtin len not in function call
// line 9: use of builtin len not in function call
// line 10: use of builtin len not in function call
// line 11: use of builtin len not in function call
// line 12: use of builtin len not in function call

// GoTypeCheckError:
// line 8
// line 9
// line 10
// line 11
// line 12

// Clean: errorcheck

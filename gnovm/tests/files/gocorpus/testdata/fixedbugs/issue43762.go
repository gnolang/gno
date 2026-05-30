// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ = true == '\\' // ERROR "invalid operation: (cannot compare true)|(true) == '\\\\' \(mismatched types untyped bool and untyped rune\)"
var _ = true == '\'' // ERROR "invalid operation: (cannot compare true)|(true) == '\\'' \(mismatched types untyped bool and untyped rune\)"
var _ = true == '\n' // ERROR "invalid operation: (cannot compare true)|(true) == '\\n' \(mismatched types untyped bool and untyped rune\)"

// GnoError:
// line 9: invalid operation: (mismatched types <untyped> bool and <untyped> int32)
// line 10: invalid operation: (mismatched types <untyped> bool and <untyped> int32)
// line 11: invalid operation: (mismatched types <untyped> bool and <untyped> int32)

// GoTypeCheckError:
// line 9: invalid operation: true == '\\' (mismatched types untyped bool and untyped rune)
// line 10: invalid operation: true == '\'' (mismatched types untyped bool and untyped rune)
// line 11: invalid operation: true == '\n' (mismatched types untyped bool and untyped rune)

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package issue42695

const _ = 6e5518446744 // ERROR "malformed constant"
const _ uint8 = 6e5518446744 // ERROR "malformed constant"

var _ = 6e5518446744 // ERROR "malformed constant"
var _ uint8 = 6e5518446744 // ERROR "malformed constant"

func f(x int) int {
        return x + 6e5518446744 // ERROR "malformed constant"
}

var _ = f(6e5518446744 /* ERROR "malformed constant" */ )

// GnoError:
// line 7: invalid decimal constant: 6e5518446744
// line 8: invalid decimal constant: 6e5518446744
// line 10: invalid decimal constant: 6e5518446744
// line 11: invalid decimal constant: 6e5518446744
// line 14: invalid decimal constant: 6e5518446744
// line 17: invalid decimal constant: 6e5518446744

// GoTypeCheckError:
// line 7
// line 8
// line 10
// line 11
// line 14
// line 17

// Clean: errorcheck

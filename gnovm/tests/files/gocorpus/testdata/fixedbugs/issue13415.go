// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that error message regarding := appears on
// correct line (and not on the line of the 2nd :=).

package p

func f() {
    select {
    case x, x := <-func() chan int { // ERROR "x repeated on left side of :=|redefinition|declared and not used"
            c := make(chan int)
            return c
    }():
    }
}

// GnoError:
// line 13: select statements are not permitted
// line 14: expected '}', found 'case' (and 1 more errors)
// line 15: channels are not permitted
// line 17: expected ';', found '('
// line 19: expected declaration, found '}'

// GoTypeCheckError:
// line 14: x repeated on left side of :=

// GnoOverStrictError:
// line 13: select statements are not permitted
// line 15: channels are not permitted
// line 17: expected ';', found '('
// line 19: expected declaration, found '}'

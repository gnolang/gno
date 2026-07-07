// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that erroneous labels are caught by the compiler.
// This set is caught by pass 2. That's why this file is label1.go.
// Does not compile.

package main

var x int

func f1() {
	switch x {
	case 1:
		continue // ERROR "continue is not in a loop$|continue statement not within for"
	}
	select {
	default:
		continue // ERROR "continue is not in a loop$|continue statement not within for"
	}

}

func f2() {
L1:
	for {
		if x == 0 {
			break L1
		}
		if x == 1 {
			continue L1
		}
		goto L1
	}

L2:
	select {
	default:
		if x == 0 {
			break L2
		}
		if x == 1 {
			continue L2 // ERROR "invalid continue label .*L2|continue is not in a loop$"
		}
		goto L2
	}

	for {
		if x == 1 {
			continue L2 // ERROR "invalid continue label .*L2"
		}
	}

L3:
	switch {
	case x > 10:
		if x == 11 {
			break L3
		}
		if x == 12 {
			continue L3 // ERROR "invalid continue label .*L3|continue is not in a loop$"
		}
		goto L3
	}

L4:
	if true {
		if x == 13 {
			break L4 // ERROR "invalid break label .*L4"
		}
		if x == 14 {
			continue L4 // ERROR "invalid continue label .*L4|continue is not in a loop$"
		}
		if x == 15 {
			goto L4
		}
	}

L5:
	f2()
	if x == 16 {
		break L5 // ERROR "invalid break label .*L5"
	}
	if x == 17 {
		continue L5 // ERROR "invalid continue label .*L5|continue is not in a loop$"
	}
	if x == 18 {
		goto L5
	}

	for {
		if x == 19 {
			break L1 // ERROR "invalid break label .*L1"
		}
		if x == 20 {
			continue L1 // ERROR "invalid continue label .*L1"
		}
		if x == 21 {
			goto L1
		}
	}

	continue // ERROR "continue is not in a loop$|continue statement not within for"
	for {
		continue on // ERROR "continue label not defined: on|invalid continue label .*on"
	}

	break // ERROR "break is not in a loop, switch, or select|break statement not within for or switch or select"
	for {
		break dance // ERROR "break label not defined: dance|invalid break label .*dance"
	}

	for {
		switch x {
		case 1:
			continue
		}
	}
}

// GnoError:
// line 20: select statements are not permitted
// line 21: expected '}', found 'default' (and 1 more errors)
// line 25: expected declaration, found '}'
// line 40: select statements are not permitted
// line 41: expected statement, found 'default' (and 1 more errors)
// line 51: expected declaration, found 'for'
// line 52: expected declaration, found 'if'
// line 53: expected declaration, found 'continue'
// line 54: expected declaration, found '}'
// line 55: expected declaration, found '}'
// line 57: expected declaration, found L3
// line 58: expected declaration, found 'switch'
// line 59: expected declaration, found 'case'
// line 60: expected declaration, found 'if'
// line 61: expected declaration, found 'break'
// line 62: expected declaration, found '}'
// line 63: expected declaration, found 'if'
// line 64: expected declaration, found 'continue'
// line 65: expected declaration, found '}'
// line 66: expected declaration, found 'goto'
// line 67: expected declaration, found '}'
// line 69: expected declaration, found L4
// line 70: expected declaration, found 'if'
// line 71: expected declaration, found 'if'
// line 72: expected declaration, found 'break'
// line 73: expected declaration, found '}'
// line 74: expected declaration, found 'if'
// line 75: expected declaration, found 'continue'
// line 76: expected declaration, found '}'
// line 77: expected declaration, found 'if'
// line 78: expected declaration, found 'goto'
// line 79: expected declaration, found '}'
// line 80: expected declaration, found '}'
// line 82: expected declaration, found L5
// line 83: expected declaration, found f2
// line 84: expected declaration, found 'if'
// line 108: label on undefined (and 1 more errors)
// line 113: label dance undefined

// GoTypeCheckError:
// line 18: continue not in for statement
// line 22: continue not in for statement
// line 46: invalid continue label L2
// line 53: invalid continue label L2
// line 64: invalid continue label L3
// line 72: invalid break label L4
// line 75: invalid continue label L4
// line 85: invalid break label L5
// line 88: invalid continue label L5
// line 96: invalid break label L1
// line 99: invalid continue label L1
// line 106: continue not in for statement
// line 108: invalid continue label on
// line 111: break not in for, switch, or select statement
// line 113: invalid break label dance

// GnoOverStrictError:
// line 20: select statements are not permitted
// line 21: expected '}', found 'default' (and 1 more errors)
// line 25: expected declaration, found '}'
// line 40: select statements are not permitted
// line 41: expected statement, found 'default' (and 1 more errors)
// line 51: expected declaration, found 'for'
// line 52: expected declaration, found 'if'
// line 54: expected declaration, found '}'
// line 55: expected declaration, found '}'
// line 57: expected declaration, found L3
// line 58: expected declaration, found 'switch'
// line 59: expected declaration, found 'case'
// line 60: expected declaration, found 'if'
// line 61: expected declaration, found 'break'
// line 62: expected declaration, found '}'
// line 63: expected declaration, found 'if'
// line 65: expected declaration, found '}'
// line 66: expected declaration, found 'goto'
// line 67: expected declaration, found '}'
// line 69: expected declaration, found L4
// line 70: expected declaration, found 'if'
// line 71: expected declaration, found 'if'
// line 73: expected declaration, found '}'
// line 74: expected declaration, found 'if'
// line 76: expected declaration, found '}'
// line 77: expected declaration, found 'if'
// line 78: expected declaration, found 'goto'
// line 79: expected declaration, found '}'
// line 80: expected declaration, found '}'
// line 82: expected declaration, found L5
// line 83: expected declaration, found f2
// line 84: expected declaration, found 'if'

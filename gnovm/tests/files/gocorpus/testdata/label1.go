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
// line 108: label on undefined
// line 113: label dance undefined

// ParseBlocked:
// line 18: gc expects: continue is not in a loop$|continue statement not within for
// line 22: gc expects: continue is not in a loop$|continue statement not within for
// line 46: gc expects: invalid continue label .*L2|continue is not in a loop$
// line 53: gc expects: invalid continue label .*L2
// line 64: gc expects: invalid continue label .*L3|continue is not in a loop$
// line 72: gc expects: invalid break label .*L4
// line 75: gc expects: invalid continue label .*L4|continue is not in a loop$
// line 85: gc expects: invalid break label .*L5
// line 88: gc expects: invalid continue label .*L5|continue is not in a loop$
// line 96: gc expects: invalid break label .*L1
// line 99: gc expects: invalid continue label .*L1
// line 106: gc expects: continue is not in a loop$|continue statement not within for
// line 111: gc expects: break is not in a loop, switch, or select|break statement not within for or switch or select

// GnoWordingDiffers:
// line 108: gc expects: continue label not defined: on|invalid continue label .*on; gno: label on undefined
// line 113: gc expects: break label not defined: dance|invalid break label .*dance; gno: label dance undefined

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 18, 22, 46, 53, 64, 72, 75, 85, 88, 96, 99, 106, 111

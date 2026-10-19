// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that pointers and interface types cannot be method receivers.
// Does not compile.

package main

type T struct {
	a int
}
type P *T
type P1 *T

func (p P) val() int   { return 1 } // ERROR "receiver.* pointer|invalid pointer or interface receiver|invalid receiver"
func (p *P1) val() int { return 1 } // ERROR "receiver.* pointer|invalid pointer or interface receiver|invalid receiver"

type I interface{}
type I1 interface{}

func (p I) val() int   { return 1 } // ERROR "receiver.*interface|invalid pointer or interface receiver"
func (p *I1) val() int { return 1 } // ERROR "receiver.*interface|invalid pointer or interface receiver"

type Val interface {
	val() int
}

var _ = (*Val).val // ERROR "method|type \*Val is pointer to interface, not interface"

var v Val
var pv = &v

var _ = pv.val() // ERROR "undefined|pointer to interface"
var _ = pv.val   // ERROR "undefined|pointer to interface"

func (t *T) g() int { return t.a }

var _ = (T).g() // ERROR "needs pointer receiver|undefined|method requires pointer|cannot call pointer method"

// GnoError:
// line 18: invalid receiver type main.P (base type is pointer type)
// line 19: invalid receiver type *main.P1 (base type is pointer type)
// line 24: invalid receiver type main.I (base type is interface type)
// line 25: invalid receiver type *main.I1 (base type is interface type)
// line 31: (*Val).val undefined (type *main.Val is pointer to interface, not interface)
// line 36: should not happen
// line 37: should not happen
// line 41: invalid method expression T.g (needs pointer receiver (*T).g)

// GnoWordingDiffers:
// line 31: gc expects: method|type \*Val is pointer to interface, not interface; gno: (*Val).val undefined (type *main.Val is pointer to interface, not interface)

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 36, 37; internal errors: lines 36, 37

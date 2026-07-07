// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Used to crash compiler in interface type equality check.
// (This test used to have problems - see #15596.)

package p

// exported interfaces

type I1 interface { // ERROR "invalid recursive type: anonymous interface refers to itself"
      F() interface{I1}
}

type I2 interface { // ERROR "invalid recursive type: anonymous interface refers to itself"
      F() interface{I2}
}

var V1 I1
var V2 I2

func F() bool {
       return V1 == V2
}

// non-exported interfaces

type i1 interface { // ERROR "invalid recursive type: anonymous interface refers to itself"
      F() interface{i1}
}

type i2 interface { // ERROR "invalid recursive type: anonymous interface refers to itself"
      F() interface{i2}
}

var v1 i1
var v2 i2

func f() bool {
       return v1 == v2
}

// GnoError:
// line 14: 2: invalid recursive type: I1 -> I1
// line 15: expected declaration, found F
// line 16: expected declaration, found '}'
// line 18: 2: invalid recursive type: I2 -> I2
// line 19: expected declaration, found F
// line 20: expected declaration, found '}'
// line 31: 2: invalid recursive type: i1 -> i1
// line 32: expected declaration, found F
// line 33: expected declaration, found '}'
// line 35: 2: invalid recursive type: i2 -> i2
// line 36: expected declaration, found F
// line 37: expected declaration, found '}'

// GoTypeCheckError:
// line 14: undefined: V1

// GnoOverStrictError:
// line 15: expected declaration, found F
// line 16: expected declaration, found '}'
// line 19: expected declaration, found F
// line 20: expected declaration, found '}'
// line 32: expected declaration, found F
// line 33: expected declaration, found '}'
// line 36: expected declaration, found F
// line 37: expected declaration, found '}'

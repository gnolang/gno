// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

const (
	zero = iota
	one
	two
	three
)

const iii int = 0x3

func f(v int) {
	switch v {
	case zero, one:
	case two, one: // ERROR "previous case at LINE-1|duplicate case .*in.* switch"

	case three:
	case 3: // ERROR "previous case at LINE-1|duplicate case .*in.* switch"
	case iii: // ERROR "previous case at LINE-2|duplicate case .*in.* switch"
	}
}

const b = "b"

var _ = map[string]int{
	"a": 0,
	b:   1,
	"a": 2, // ERROR "previous key at LINE-2|duplicate key.*in map literal"
	"b": 3, // GC_ERROR "previous key at LINE-2|duplicate key.*in map literal"
	"b": 4, // GC_ERROR "previous key at LINE-3|duplicate key.*in map literal"
}

// GnoError:
// line 31: 2: duplicate key ("a" string) in map literal
// line 32: expected declaration, found "a"
// line 33: expected declaration, found b
// line 34: expected declaration, found "a"
// line 35: expected declaration, found "b"
// line 36: expected declaration, found "b"
// line 37: expected declaration, found '}'

// GoTypeCheckError:
// line 21: duplicate case one (constant 1 of type int) in expression switch
// line 24: duplicate case 3 (constant of type int) in expression switch
// line 25: duplicate case iii (constant 3 of type int) in expression switch
// line 34: duplicate key "a" in map literal
// line 35: duplicate key "b" in map literal
// line 36: duplicate key "b" in map literal

// GnoOverStrictError:
// line 31: 2: duplicate key ("a" string) in map literal
// line 32: expected declaration, found "a"
// line 33: expected declaration, found b
// line 37: expected declaration, found '}'

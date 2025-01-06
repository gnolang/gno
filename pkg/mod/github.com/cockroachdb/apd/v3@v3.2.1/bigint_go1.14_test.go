// Copyright 2022 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

//go:build go1.14
// +build go1.14

package apd

import (
	"testing"
	"testing/quick"
)

//////////////////////////////////////////////////////////////////////////////////
// The following tests were copied from the standard library's math/big package //
//////////////////////////////////////////////////////////////////////////////////

func checkGcd(aBytes, bBytes []byte) bool {
	x := new(BigInt)
	y := new(BigInt)
	a := new(BigInt).SetBytes(aBytes)
	b := new(BigInt).SetBytes(bBytes)

	d := new(BigInt).GCD(x, y, a, b)
	x.Mul(x, a)
	y.Mul(y, b)
	x.Add(x, y)

	return x.Cmp(d) == 0
}

var gcdTests = []struct {
	d, x, y, a, b string
}{
	// a <= 0 || b <= 0
	{"0", "0", "0", "0", "0"},
	{"7", "0", "1", "0", "7"},
	{"7", "0", "-1", "0", "-7"},
	{"11", "1", "0", "11", "0"},
	{"7", "-1", "-2", "-77", "35"},
	{"935", "-3", "8", "64515", "24310"},
	{"935", "-3", "-8", "64515", "-24310"},
	{"935", "3", "-8", "-64515", "-24310"},

	{"1", "-9", "47", "120", "23"},
	{"7", "1", "-2", "77", "35"},
	{"935", "-3", "8", "64515", "24310"},
	{"935000000000000000", "-3", "8", "64515000000000000000", "24310000000000000000"},
	{"1", "-221", "22059940471369027483332068679400581064239780177629666810348940098015901108344", "98920366548084643601728869055592650835572950932266967461790948584315647051443", "991"},
}

func testGcd(t *testing.T, d, x, y, a, b *BigInt) {
	var X *BigInt
	if x != nil {
		X = new(BigInt)
	}
	var Y *BigInt
	if y != nil {
		Y = new(BigInt)
	}

	D := new(BigInt).GCD(X, Y, a, b)
	if D.Cmp(d) != 0 {
		t.Errorf("GCD(%s, %s, %s, %s): got d = %s, want %s", x, y, a, b, D, d)
	}
	if x != nil && X.Cmp(x) != 0 {
		t.Errorf("GCD(%s, %s, %s, %s): got x = %s, want %s", x, y, a, b, X, x)
	}
	if y != nil && Y.Cmp(y) != 0 {
		t.Errorf("GCD(%s, %s, %s, %s): got y = %s, want %s", x, y, a, b, Y, y)
	}

	// check results in presence of aliasing (issue #11284)
	a2 := new(BigInt).Set(a)
	b2 := new(BigInt).Set(b)
	a2.GCD(X, Y, a2, b2) // result is same as 1st argument
	if a2.Cmp(d) != 0 {
		t.Errorf("aliased z = a GCD(%s, %s, %s, %s): got d = %s, want %s", x, y, a, b, a2, d)
	}
	if x != nil && X.Cmp(x) != 0 {
		t.Errorf("aliased z = a GCD(%s, %s, %s, %s): got x = %s, want %s", x, y, a, b, X, x)
	}
	if y != nil && Y.Cmp(y) != 0 {
		t.Errorf("aliased z = a GCD(%s, %s, %s, %s): got y = %s, want %s", x, y, a, b, Y, y)
	}

	a2 = new(BigInt).Set(a)
	b2 = new(BigInt).Set(b)
	b2.GCD(X, Y, a2, b2) // result is same as 2nd argument
	if b2.Cmp(d) != 0 {
		t.Errorf("aliased z = b GCD(%s, %s, %s, %s): got d = %s, want %s", x, y, a, b, b2, d)
	}
	if x != nil && X.Cmp(x) != 0 {
		t.Errorf("aliased z = b GCD(%s, %s, %s, %s): got x = %s, want %s", x, y, a, b, X, x)
	}
	if y != nil && Y.Cmp(y) != 0 {
		t.Errorf("aliased z = b GCD(%s, %s, %s, %s): got y = %s, want %s", x, y, a, b, Y, y)
	}

	a2 = new(BigInt).Set(a)
	b2 = new(BigInt).Set(b)
	D = new(BigInt).GCD(a2, b2, a2, b2) // x = a, y = b
	if D.Cmp(d) != 0 {
		t.Errorf("aliased x = a, y = b GCD(%s, %s, %s, %s): got d = %s, want %s", x, y, a, b, D, d)
	}
	if x != nil && a2.Cmp(x) != 0 {
		t.Errorf("aliased x = a, y = b GCD(%s, %s, %s, %s): got x = %s, want %s", x, y, a, b, a2, x)
	}
	if y != nil && b2.Cmp(y) != 0 {
		t.Errorf("aliased x = a, y = b GCD(%s, %s, %s, %s): got y = %s, want %s", x, y, a, b, b2, y)
	}

	a2 = new(BigInt).Set(a)
	b2 = new(BigInt).Set(b)
	D = new(BigInt).GCD(b2, a2, a2, b2) // x = b, y = a
	if D.Cmp(d) != 0 {
		t.Errorf("aliased x = b, y = a GCD(%s, %s, %s, %s): got d = %s, want %s", x, y, a, b, D, d)
	}
	if x != nil && b2.Cmp(x) != 0 {
		t.Errorf("aliased x = b, y = a GCD(%s, %s, %s, %s): got x = %s, want %s", x, y, a, b, b2, x)
	}
	if y != nil && a2.Cmp(y) != 0 {
		t.Errorf("aliased x = b, y = a GCD(%s, %s, %s, %s): got y = %s, want %s", x, y, a, b, a2, y)
	}
}

// This was not supported in go1.13. See https://go.dev/doc/go1.14:
// > The GCD method now allows the inputs a and b to be zero or negative.
func TestBigIntGcd(t *testing.T) {
	for _, test := range gcdTests {
		d, _ := new(BigInt).SetString(test.d, 0)
		x, _ := new(BigInt).SetString(test.x, 0)
		y, _ := new(BigInt).SetString(test.y, 0)
		a, _ := new(BigInt).SetString(test.a, 0)
		b, _ := new(BigInt).SetString(test.b, 0)

		testGcd(t, d, nil, nil, a, b)
		testGcd(t, d, x, nil, a, b)
		testGcd(t, d, nil, y, a, b)
		testGcd(t, d, x, y, a, b)
	}

	if err := quick.Check(checkGcd, nil); err != nil {
		t.Error(err)
	}
}

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// GNO modifications:
// This file has been modified to remove the dependency on "testing/quick" package, which
// is not available in gno due to its reliance on refledtion. The original file can be found here:
// https://cs.opensource.google/go/go/+/master:src/crypto/subtle/constant_time_test.go

package subtle

import (
	"bytes"
	"testing"
)

type TestConstantTimeCompareStruct struct {
	a, b []byte
	out  int
}

var testConstantTimeCompareData = []TestConstantTimeCompareStruct{
	{[]byte{}, []byte{}, 1},
	{[]byte{0x11}, []byte{0x11}, 1},
	{[]byte{0x12}, []byte{0x11}, 0},
	{[]byte{0x11}, []byte{0x11, 0x12}, 0},
	{[]byte{0x11, 0x12}, []byte{0x11}, 0},
}

func TestConstantTimeCompare(t *testing.T) {
	for i, test := range testConstantTimeCompareData {
		if r := ConstantTimeCompare(test.a, test.b); r != test.out {
			t.Errorf("#%d bad result (got %x, want %x)", i, r, test.out)
		}
	}
}

type TestConstantTimeByteEqStruct struct {
	a, b uint8
	out  int
}

var testConstandTimeByteEqData = []TestConstantTimeByteEqStruct{
	{0, 0, 1},
	{0, 1, 0},
	{1, 0, 0},
	{0xff, 0xff, 1},
	{0xff, 0xfe, 0},
}

func byteEq(a, b uint8) int {
	if a == b {
		return 1
	}
	return 0
}

func TestConstantTimeByteEq(t *testing.T) {
	for i, test := range testConstandTimeByteEqData {
		if r := ConstantTimeByteEq(test.a, test.b); r != test.out {
			t.Errorf("#%d bad result (got %x, want %x)", i, r, test.out)
		}
	}

	tt := []struct {
		a, b uint8
	}{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
		{0x7F, 0x7F},
		{0x7F, 0},
		{0, 0x7F},
	}

	for i, test := range tt {
		if r := ConstantTimeByteEq(test.a, test.b); r != byteEq(test.a, test.b) {
			t.Errorf("#%d bad result (got %x, want %x)", i, r, byteEq(test.a, test.b))
		}
	}
}

func eq(a, b int32) int {
	if a == b {
		return 1
	}
	return 0
}

func TestConstantTimeEq(t *testing.T) {
	tt := []struct {
		a, b int32
	}{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
		{400, 400},
		{400, 401},
		{401, 400},
		{0x7FFFFFFF, 0x7FFFFFFF},
		{0x7FFFFFFF, 0},
		{0, 0x7FFFFFFF},
	}

	for i, test := range tt {
		if r := ConstantTimeEq(test.a, test.b); r != eq(test.a, test.b) {
			t.Errorf("#%d bad result (got %x, want %x)", i, r, eq(test.a, test.b))
		}
	}
}

func makeCopy(v int, x, y []byte) []byte {
	if len(x) > len(y) {
		x = x[:len(y)]
	} else {
		y = y[:len(x)]
	}
	if v == 1 {
		copy(x, y)
	}
	return x
}

func constantTimeCopyWrapper(v int, x, y []byte) []byte {
	if len(x) > len(y) {
		x = x[:len(y)]
	} else {
		y = y[:len(x)]
	}
	v &= 1
	ConstantTimeCopy(v, x, y)
	return x
}

func TestConstantTimeCopy(t *testing.T) {
	tt := []struct {
		v    int
		x, y []byte
	}{
		{0, []byte{1, 2, 3}, []byte{4, 5, 6}},
		{1, []byte{1, 2, 3}, []byte{4, 5, 6}},
		{0, []byte{1, 2, 3}, make([]byte, 3)},
		{1, []byte{1, 2, 3}, []byte{4, 5, 6}},
		{0, []byte{1, 2, 3}, []byte{4, 5}},
		{1, []byte{1, 2, 3}, []byte{4, 5}},
		{0, []byte{1, 2}, []byte{4, 5, 6}},
		{1, []byte{1, 2}, []byte{4, 5, 6}},
		{0, []byte{1, 2}, []byte{4, 5}},
		{1, []byte{1, 2}, []byte{4, 5}},
		{0, []byte{1, 2, 3}, []byte{4, 5}},
		{1, []byte{1, 2, 3}, []byte{4, 5}},
		{0, []byte{1, 2}, []byte{4, 5, 6}},
		{1, []byte{1, 2}, []byte{4, 5, 6}},
		{0, []byte{1, 2}, []byte{4, 5}},
		{1, []byte{1, 2}, []byte{4, 5}},
	}

	for i, test := range tt {
		r := constantTimeCopyWrapper(test.v, test.x, test.y)
		expected := makeCopy(test.v, test.x, test.y)
		if !bytes.Equal(r, expected) {
			t.Errorf("#%d bad result (got %x, want %x)", i, r, expected)
		}
	}
}

var lessOrEqTests = []struct {
	x, y, result int
}{
	{0, 0, 1},
	{1, 0, 0},
	{0, 1, 1},
	{10, 20, 1},
	{20, 10, 0},
	{10, 10, 1},
}

func TestConstantTimeLessOrEq(t *testing.T) {
	for i, test := range lessOrEqTests {
		result := ConstantTimeLessOrEq(test.x, test.y)
		if result != test.result {
			t.Errorf("#%d: %d <= %d gave %d, expected %d", i, test.x, test.y, result, test.result)
		}
	}
}

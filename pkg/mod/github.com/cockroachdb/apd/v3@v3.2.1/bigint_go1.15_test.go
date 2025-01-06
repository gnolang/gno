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

//go:build go1.15
// +build go1.15

package apd

import (
	"testing"
	"testing/quick"
)

// TestBigIntMatchesMathBigInt15 is like TestBigIntMatchesMathBigInt, but for
// parts of the shared BigInt/big.Int API that were introduced in go1.15.
func TestBigIntMatchesMathBigInt15(t *testing.T) {
	t.Run("FillBytes", func(t *testing.T) {
		apd := func(z number) []byte {
			return z.toApd(t).FillBytes(make([]byte, len(z)))
		}
		math := func(z number) []byte {
			return z.toMath(t).FillBytes(make([]byte, len(z)))
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})
}

//////////////////////////////////////////////////////////////////////////////////
// The following tests were copied from the standard library's math/big package //
//////////////////////////////////////////////////////////////////////////////////

func TestBigIntFillBytes(t *testing.T) {
	checkResult := func(t *testing.T, buf []byte, want *BigInt) {
		t.Helper()
		got := new(BigInt).SetBytes(buf)
		if got.CmpAbs(want) != 0 {
			t.Errorf("got 0x%x, want 0x%x: %x", got, want, buf)
		}
	}
	panics := func(f func()) (panic bool) {
		defer func() { panic = recover() != nil }()
		f()
		return
	}

	for _, n := range []string{
		"0",
		"1000",
		"0xffffffff",
		"-0xffffffff",
		"0xffffffffffffffff",
		"0x10000000000000000",
		"0xabababababababababababababababababababababababababa",
		"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	} {
		t.Run(n, func(t *testing.T) {
			t.Logf(n)
			x, ok := new(BigInt).SetString(n, 0)
			if !ok {
				panic("invalid test entry")
			}

			// Perfectly sized buffer.
			byteLen := (x.BitLen() + 7) / 8
			buf := make([]byte, byteLen)
			checkResult(t, x.FillBytes(buf), x)

			// Way larger, checking all bytes get zeroed.
			buf = make([]byte, 100)
			for i := range buf {
				buf[i] = 0xff
			}
			checkResult(t, x.FillBytes(buf), x)

			// Too small.
			if byteLen > 0 {
				buf = make([]byte, byteLen-1)
				if !panics(func() { x.FillBytes(buf) }) {
					t.Errorf("expected panic for small buffer and value %x", x)
				}
			}
		})
	}
}

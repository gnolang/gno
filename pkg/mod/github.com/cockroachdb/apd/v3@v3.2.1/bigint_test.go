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

package apd

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/quick"
)

// TestBigIntMatchesMathBigInt uses testing/quick to verify that all methods on
// apd.BigInt and math/big.Int have identical behavior for all inputs.
func TestBigIntMatchesMathBigInt(t *testing.T) {
	// Catch specific panics and map to strings.
	const (
		panicDivisionByZero          = "division by zero"
		panicJacobi                  = "invalid 2nd argument to Int.Jacobi: need odd integer"
		panicNegativeBit             = "negative bit index"
		panicSquareRootOfNegativeNum = "square root of negative number"
	)
	catchPanic := func(fn func() string, catches ...string) (res string) {
		defer func() {
			if r := recover(); r != nil {
				if rs, ok := r.(string); ok {
					for _, catch := range catches {
						if strings.Contains(rs, catch) {
							res = fmt.Sprintf("caught: %s", r)
						}
					}
				}
				if res == "" { // not caught
					panic(r)
				}
			}
		}()
		return fn()
	}

	t.Run("Abs", func(t *testing.T) {
		apd := func(z, x number) string {
			return z.toApd(t).Abs(x.toApd(t)).String()
		}
		math := func(z, x number) string {
			return z.toMath(t).Abs(x.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Add", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return z.toApd(t).Add(x.toApd(t), y.toApd(t)).String()
		}
		math := func(z, x, y number) string {
			return z.toMath(t).Add(x.toMath(t), y.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("And", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return z.toApd(t).And(x.toApd(t), y.toApd(t)).String()
		}
		math := func(z, x, y number) string {
			return z.toMath(t).And(x.toMath(t), y.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("AndNot", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return z.toApd(t).AndNot(x.toApd(t), y.toApd(t)).String()
		}
		math := func(z, x, y number) string {
			return z.toMath(t).AndNot(x.toMath(t), y.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Append", func(t *testing.T) {
		apd := func(z numberOrNil) []byte {
			return z.toApd(t).Append(nil, 10)
		}
		math := func(z numberOrNil) []byte {
			return z.toMath(t).Append(nil, 10)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Binomial", func(t *testing.T) {
		t.Skip("too slow")
		apd := func(z number, n, k int64) string {
			return z.toApd(t).Binomial(n, k).String()
		}
		math := func(z number, n, k int64) string {
			return z.toMath(t).Binomial(n, k).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Bit", func(t *testing.T) {
		apd := func(z number, i int) string {
			return catchPanic(func() string {
				return strconv.FormatUint(uint64(z.toApd(t).Bit(i)), 10)
			}, panicNegativeBit)
		}
		math := func(z number, i int) string {
			return catchPanic(func() string {
				return strconv.FormatUint(uint64(z.toMath(t).Bit(i)), 10)
			}, panicNegativeBit)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("BitLen", func(t *testing.T) {
		apd := func(z number) int {
			return z.toApd(t).BitLen()
		}
		math := func(z number) int {
			return z.toMath(t).BitLen()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Bits", func(t *testing.T) {
		emptyToNil := func(w []big.Word) []big.Word {
			if len(w) == 0 {
				return nil
			}
			return w
		}
		apd := func(z number) []big.Word {
			return emptyToNil(z.toApd(t).Bits())
		}
		math := func(z number) []big.Word {
			return emptyToNil(z.toMath(t).Bits())
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Bytes", func(t *testing.T) {
		apd := func(z number) []byte {
			return z.toApd(t).Bytes()
		}
		math := func(z number) []byte {
			return z.toMath(t).Bytes()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Cmp", func(t *testing.T) {
		apd := func(z, y number) int {
			return z.toApd(t).Cmp(y.toApd(t))
		}
		math := func(z, y number) int {
			return z.toMath(t).Cmp(y.toMath(t))
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("CmpAbs", func(t *testing.T) {
		apd := func(z, y number) int {
			return z.toApd(t).CmpAbs(y.toApd(t))
		}
		math := func(z, y number) int {
			return z.toMath(t).CmpAbs(y.toMath(t))
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Div", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toApd(t).Div(x.toApd(t), y.toApd(t)).String()
			}, panicDivisionByZero)
		}
		math := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toMath(t).Div(x.toMath(t), y.toMath(t)).String()
			}, panicDivisionByZero)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("DivMod", func(t *testing.T) {
		apd := func(z, x, y, m number) string {
			return catchPanic(func() string {
				zi, mi := z.toApd(t), m.toApd(t)
				zi.DivMod(x.toApd(t), y.toApd(t), mi)
				return zi.String() + " | " + mi.String()
			}, panicDivisionByZero)
		}
		math := func(z, x, y, m number) string {
			return catchPanic(func() string {
				zi, mi := z.toMath(t), m.toMath(t)
				zi.DivMod(x.toMath(t), y.toMath(t), mi)
				return zi.String() + " | " + mi.String()
			}, panicDivisionByZero)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Exp", func(t *testing.T) {
		t.Skip("too slow")
		apd := func(z, x, y, m number) string {
			return z.toApd(t).Exp(x.toApd(t), y.toApd(t), m.toApd(t)).String()
		}
		math := func(z, x, y, m number) string {
			return z.toMath(t).Exp(x.toMath(t), y.toMath(t), m.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Format", func(t *testing.T) {
		// Call indirectly through fmt.Sprint.
		apd := func(z numberOrNil) string {
			return fmt.Sprint(z.toApd(t))
		}
		math := func(z numberOrNil) string {
			return fmt.Sprint(z.toMath(t))
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("GCD", func(t *testing.T) {
		apd := func(z number, x, y numberOrNil, a, b number) string {
			return z.toApd(t).GCD(x.toApd(t), y.toApd(t), a.toApd(t), b.toApd(t)).String()
		}
		math := func(z number, x, y numberOrNil, a, b number) string {
			return z.toMath(t).GCD(x.toMath(t), y.toMath(t), a.toMath(t), b.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("GobEncode", func(t *testing.T) {
		apd := func(z numberOrNil) ([]byte, error) {
			return z.toApd(t).GobEncode()
		}
		math := func(z numberOrNil) ([]byte, error) {
			return z.toMath(t).GobEncode()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("GobDecode", func(t *testing.T) {
		apd := func(z number, buf []byte) (string, error) {
			zi := z.toApd(t)
			err := zi.GobDecode(buf)
			return zi.String(), err
		}
		math := func(z number, buf []byte) (string, error) {
			zi := z.toMath(t)
			err := zi.GobDecode(buf)
			return zi.String(), err
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Int64", func(t *testing.T) {
		apd := func(z number) int64 {
			return z.toApd(t).Int64()
		}
		math := func(z number) int64 {
			return z.toMath(t).Int64()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("IsInt64", func(t *testing.T) {
		apd := func(z number) bool {
			return z.toApd(t).IsInt64()
		}
		math := func(z number) bool {
			return z.toMath(t).IsInt64()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("IsUint64", func(t *testing.T) {
		apd := func(z number) bool {
			return z.toApd(t).IsUint64()
		}
		math := func(z number) bool {
			return z.toMath(t).IsUint64()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Lsh", func(t *testing.T) {
		const maxShift = 1000 // avoid makeslice: len out of range
		apd := func(z, x number, n uint) string {
			if n > maxShift {
				n = maxShift
			}
			return z.toApd(t).Lsh(x.toApd(t), n).String()
		}
		math := func(z, x number, n uint) string {
			if n > maxShift {
				n = maxShift
			}
			return z.toMath(t).Lsh(x.toMath(t), n).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		apd := func(z numberOrNil) ([]byte, error) {
			return z.toApd(t).MarshalJSON()
		}
		math := func(z numberOrNil) ([]byte, error) {
			return z.toMath(t).MarshalJSON()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("MarshalText", func(t *testing.T) {
		apd := func(z numberOrNil) ([]byte, error) {
			return z.toApd(t).MarshalText()
		}
		math := func(z numberOrNil) ([]byte, error) {
			return z.toMath(t).MarshalText()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Mod", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toApd(t).Mod(x.toApd(t), y.toApd(t)).String()
			}, panicDivisionByZero, panicJacobi)
		}
		math := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toMath(t).Mod(x.toMath(t), y.toMath(t)).String()
			}, panicDivisionByZero, panicJacobi)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("ModInverse", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toApd(t).ModInverse(x.toApd(t), y.toApd(t)).String()
			}, panicDivisionByZero)
		}
		math := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toMath(t).ModInverse(x.toMath(t), y.toMath(t)).String()
			}, panicDivisionByZero)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("ModSqrt", func(t *testing.T) {
		t.Skip("too slow")
		apd := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toApd(t).ModSqrt(x.toApd(t), y.toApd(t)).String()
			}, panicJacobi)
		}
		math := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toMath(t).ModSqrt(x.toMath(t), y.toMath(t)).String()
			}, panicJacobi)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Mul", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return z.toApd(t).Mul(x.toApd(t), y.toApd(t)).String()
		}
		math := func(z, x, y number) string {
			return z.toMath(t).Mul(x.toMath(t), y.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("MulRange", func(t *testing.T) {
		t.Skip("too slow")
		apd := func(z number, x, y int64) string {
			return z.toApd(t).MulRange(x, y).String()
		}
		math := func(z number, x, y int64) string {
			return z.toMath(t).MulRange(x, y).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Neg", func(t *testing.T) {
		apd := func(z, x number) string {
			return z.toApd(t).Neg(x.toApd(t)).String()
		}
		math := func(z, x number) string {
			return z.toMath(t).Neg(x.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Not", func(t *testing.T) {
		apd := func(z, x number) string {
			return z.toApd(t).Not(x.toApd(t)).String()
		}
		math := func(z, x number) string {
			return z.toMath(t).Not(x.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Or", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return z.toApd(t).Or(x.toApd(t), y.toApd(t)).String()
		}
		math := func(z, x, y number) string {
			return z.toMath(t).Or(x.toMath(t), y.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("ProbablyPrime", func(t *testing.T) {
		apd := func(z number) bool {
			return z.toApd(t).ProbablyPrime(64)
		}
		math := func(z number) bool {
			return z.toMath(t).ProbablyPrime(64)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Quo", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toApd(t).Quo(x.toApd(t), y.toApd(t)).String()
			}, panicDivisionByZero)
		}
		math := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toMath(t).Quo(x.toMath(t), y.toMath(t)).String()
			}, panicDivisionByZero)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("QuoRem", func(t *testing.T) {
		apd := func(z, x, y, r number) string {
			return catchPanic(func() string {
				zi, ri := z.toApd(t), r.toApd(t)
				zi.QuoRem(x.toApd(t), y.toApd(t), ri)
				return zi.String() + " | " + ri.String()
			}, panicDivisionByZero)
		}
		math := func(z, x, y, r number) string {
			return catchPanic(func() string {
				zi, ri := z.toMath(t), r.toMath(t)
				zi.QuoRem(x.toMath(t), y.toMath(t), ri)
				return zi.String() + " | " + ri.String()
			}, panicDivisionByZero)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Rand", func(t *testing.T) {
		apd := func(z, n number, seed int64) string {
			rng := rand.New(rand.NewSource(seed))
			return z.toApd(t).Rand(rng, n.toApd(t)).String()
		}
		math := func(z, n number, seed int64) string {
			rng := rand.New(rand.NewSource(seed))
			return z.toMath(t).Rand(rng, n.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Rem", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toApd(t).Rem(x.toApd(t), y.toApd(t)).String()
			}, panicDivisionByZero)
		}
		math := func(z, x, y number) string {
			return catchPanic(func() string {
				return z.toMath(t).Rem(x.toMath(t), y.toMath(t)).String()
			}, panicDivisionByZero)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Rsh", func(t *testing.T) {
		const maxShift = 1000 // avoid makeslice: len out of range
		apd := func(z, x number, n uint) string {
			if n > maxShift {
				n = maxShift
			}
			return z.toApd(t).Rsh(x.toApd(t), n).String()
		}
		math := func(z, x number, n uint) string {
			if n > maxShift {
				n = maxShift
			}
			return z.toMath(t).Rsh(x.toMath(t), n).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Scan", func(t *testing.T) {
		// Call indirectly through fmt.Sscan.
		apd := func(z, src number) (string, error) {
			zi := z.toApd(t)
			_, err := fmt.Sscan(string(src), zi)
			return zi.String(), err
		}
		math := func(z, src number) (string, error) {
			zi := z.toMath(t)
			_, err := fmt.Sscan(string(src), zi)
			return zi.String(), err
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Set", func(t *testing.T) {
		apd := func(z, x number) string {
			return z.toApd(t).Set(x.toApd(t)).String()
		}
		math := func(z, x number) string {
			return z.toMath(t).Set(x.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("SetBit", func(t *testing.T) {
		const maxBit = 1000 // avoid makeslice: len out of range
		apd := func(z, x number, i int, b bool) string {
			if i > maxBit {
				i = maxBit
			}
			bi := uint(0)
			if b {
				bi = 1
			}
			return catchPanic(func() string {
				return z.toApd(t).SetBit(x.toApd(t), i, bi).String()
			}, panicNegativeBit)
		}
		math := func(z, x number, i int, b bool) string {
			if i > maxBit {
				i = maxBit
			}
			bi := uint(0)
			if b {
				bi = 1
			}
			return catchPanic(func() string {
				return z.toMath(t).SetBit(x.toMath(t), i, bi).String()
			}, panicNegativeBit)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("SetBits", func(t *testing.T) {
		apd := func(z number, abs []big.Word) string {
			return z.toApd(t).SetBits(abs).String()
		}
		math := func(z number, abs []big.Word) string {
			return z.toMath(t).SetBits(abs).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("SetBytes", func(t *testing.T) {
		apd := func(z number, buf []byte) string {
			return z.toApd(t).SetBytes(buf).String()
		}
		math := func(z number, buf []byte) string {
			return z.toMath(t).SetBytes(buf).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("SetInt64", func(t *testing.T) {
		apd := func(z number, x int64) string {
			return z.toApd(t).SetInt64(x).String()
		}
		math := func(z number, x int64) string {
			return z.toMath(t).SetInt64(x).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("SetString", func(t *testing.T) {
		apd := func(z, x number) (string, bool) {
			zi, ok := z.toApd(t).SetString(string(x), 10)
			return zi.String(), ok
		}
		math := func(z, x number) (string, bool) {
			zi, ok := z.toMath(t).SetString(string(x), 10)
			return zi.String(), ok
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("SetUint64", func(t *testing.T) {
		apd := func(z number, x uint64) string {
			return z.toApd(t).SetUint64(x).String()
		}
		math := func(z number, x uint64) string {
			return z.toMath(t).SetUint64(x).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Sign", func(t *testing.T) {
		apd := func(z number) int {
			return z.toApd(t).Sign()
		}
		math := func(z number) int {
			return z.toMath(t).Sign()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Sqrt", func(t *testing.T) {
		apd := func(z, x number) string {
			return catchPanic(func() string {
				return z.toApd(t).Sqrt(x.toApd(t)).String()
			}, panicSquareRootOfNegativeNum)
		}
		math := func(z, x number) string {
			return catchPanic(func() string {
				return z.toMath(t).Sqrt(x.toMath(t)).String()
			}, panicSquareRootOfNegativeNum)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("String", func(t *testing.T) {
		apd := func(z numberOrNil) string {
			return z.toApd(t).String()
		}
		math := func(z numberOrNil) string {
			return z.toMath(t).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Sub", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return z.toApd(t).Sub(x.toApd(t), y.toApd(t)).String()
		}
		math := func(z, x, y number) string {
			return z.toMath(t).Sub(x.toMath(t), y.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Text", func(t *testing.T) {
		apd := func(z numberOrNil) string {
			return z.toApd(t).Text(10)
		}
		math := func(z numberOrNil) string {
			return z.toMath(t).Text(10)
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("TrailingZeroBits", func(t *testing.T) {
		apd := func(z number) uint {
			return z.toApd(t).TrailingZeroBits()
		}
		math := func(z number) uint {
			return z.toMath(t).TrailingZeroBits()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Uint64", func(t *testing.T) {
		apd := func(z number) uint64 {
			return z.toApd(t).Uint64()
		}
		math := func(z number) uint64 {
			return z.toMath(t).Uint64()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		apd := func(z number, text []byte) (string, error) {
			zi := z.toApd(t)
			if err := zi.UnmarshalJSON(text); err != nil {
				return "", err
			}
			return zi.String(), nil
		}
		math := func(z number, text []byte) (string, error) {
			zi := z.toMath(t)
			if err := zi.UnmarshalJSON(text); err != nil {
				return "", err
			}
			return zi.String(), nil
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		apd := func(z number, text []byte) (string, error) {
			zi := z.toApd(t)
			if err := zi.UnmarshalText(text); err != nil {
				return "", err
			}
			return zi.String(), nil
		}
		math := func(z number, text []byte) (string, error) {
			zi := z.toMath(t)
			if err := zi.UnmarshalText(text); err != nil {
				return "", err
			}
			return zi.String(), nil
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})

	t.Run("Xor", func(t *testing.T) {
		apd := func(z, x, y number) string {
			return z.toApd(t).Xor(x.toApd(t), y.toApd(t)).String()
		}
		math := func(z, x, y number) string {
			return z.toMath(t).Xor(x.toMath(t), y.toMath(t)).String()
		}
		require(t, quick.CheckEqual(apd, math, nil))
	})
}

// TestBigIntMathBigIntRoundTrip uses testing/quick to verify that the
// apd.BigInt / math/big.Int interoperation methods each round-trip.
func TestBigIntMathBigIntRoundTrip(t *testing.T) {
	t.Run("apd->math->apd", func(t *testing.T) {
		base := func(z number) string {
			return z.toApd(t).String()
		}
		roundtrip := func(z number) string {
			bi := z.toApd(t).MathBigInt()
			return new(BigInt).SetMathBigInt(bi).String()
		}
		require(t, quick.CheckEqual(base, roundtrip, nil))
	})

	t.Run("math->apd->math", func(t *testing.T) {
		base := func(z number) string {
			return z.toMath(t).String()
		}
		roundtrip := func(z number) string {
			bi := new(BigInt).SetMathBigInt(z.toMath(t))
			return bi.MathBigInt().String()
		}
		require(t, quick.CheckEqual(base, roundtrip, nil))
	})
}

// number is a quick.Generator for large integer numbers.
type number string

func (n number) Generate(r *rand.Rand, size int) reflect.Value {
	var s string
	if r.Intn(2) != 0 {
		s = n.generateInterestingNumber(r)
	} else {
		s = n.generateRandomNumber(r, size)
	}
	return reflect.ValueOf(number(s))
}

func (z *BigInt) incr() *BigInt { return z.Add(z, bigOne) }
func (z *BigInt) decr() *BigInt { return z.Sub(z, bigOne) }

var interestingNumbers = [...]*BigInt{
	new(BigInt).SetInt64(math.MinInt64).decr(),
	new(BigInt).SetInt64(math.MinInt64),
	new(BigInt).SetInt64(math.MinInt64).incr(),
	new(BigInt).SetInt64(math.MinInt32),
	new(BigInt).SetInt64(math.MinInt16),
	new(BigInt).SetInt64(math.MinInt8),
	new(BigInt).SetInt64(0),
	new(BigInt).SetInt64(math.MaxInt8),
	new(BigInt).SetInt64(math.MaxUint8),
	new(BigInt).SetInt64(math.MaxInt16),
	new(BigInt).SetInt64(math.MaxUint16),
	new(BigInt).SetInt64(math.MaxInt32),
	new(BigInt).SetInt64(math.MaxUint32),
	new(BigInt).SetInt64(math.MaxInt64).decr(),
	new(BigInt).SetInt64(math.MaxInt64),
	new(BigInt).SetInt64(math.MaxInt64).incr(),
	new(BigInt).SetUint64(math.MaxUint64).decr(),
	new(BigInt).SetUint64(math.MaxUint64),
	new(BigInt).SetUint64(math.MaxUint64).incr(),
}

func (number) generateInterestingNumber(r *rand.Rand) string {
	return interestingNumbers[r.Intn(len(interestingNumbers))].String()
}

var numbers = [...]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

func (number) generateRandomNumber(r *rand.Rand, size int) string {
	var s strings.Builder
	if r.Intn(2) != 0 {
		s.WriteByte('-') // neg
	}
	digits := r.Intn(size) + 1
	for i := 0; i < digits; i++ {
		s.WriteByte(numbers[r.Intn(len(numbers))])
	}
	return s.String()
}

func (n number) toApd(t *testing.T) *BigInt {
	var x BigInt
	if _, ok := x.SetString(string(n), 10); !ok {
		t.Fatalf("failed to SetString(%q)", n)
	}
	return &x
}

func (n number) toMath(t *testing.T) *big.Int {
	var x big.Int
	if _, ok := x.SetString(string(n), 10); !ok {
		t.Fatalf("failed to SetString(%q)", n)
	}
	return &x
}

type numberOrNil struct {
	Num number
	Nil bool
}

func (n numberOrNil) toApd(t *testing.T) *BigInt {
	if n.Nil {
		return nil
	}
	return n.Num.toApd(t)
}

func (n numberOrNil) toMath(t *testing.T) *big.Int {
	if n.Nil {
		return nil
	}
	return n.Num.toMath(t)
}

// Until we import github.com/stretchr/testify/require.
func require(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

//////////////////////////////////////////////////////////////////////////////////
// The following tests were copied from the standard library's math/big package //
//////////////////////////////////////////////////////////////////////////////////

//
// Tests from src/math/big/int_test.go
//

type funZZ func(z, x, y *BigInt) *BigInt
type argZZ struct {
	z, x, y *BigInt
}

var sumZZ = []argZZ{
	{NewBigInt(0), NewBigInt(0), NewBigInt(0)},
	{NewBigInt(1), NewBigInt(1), NewBigInt(0)},
	{NewBigInt(1111111110), NewBigInt(123456789), NewBigInt(987654321)},
	{NewBigInt(-1), NewBigInt(-1), NewBigInt(0)},
	{NewBigInt(864197532), NewBigInt(-123456789), NewBigInt(987654321)},
	{NewBigInt(-1111111110), NewBigInt(-123456789), NewBigInt(-987654321)},
}

var prodZZ = []argZZ{
	{NewBigInt(0), NewBigInt(0), NewBigInt(0)},
	{NewBigInt(0), NewBigInt(1), NewBigInt(0)},
	{NewBigInt(1), NewBigInt(1), NewBigInt(1)},
	{NewBigInt(-991 * 991), NewBigInt(991), NewBigInt(-991)},
	// TODO(gri) add larger products
}

func TestBigIntSignZ(t *testing.T) {
	var zero BigInt
	for _, a := range sumZZ {
		s := a.z.Sign()
		e := a.z.Cmp(&zero)
		if s != e {
			t.Errorf("got %d; want %d for z = %v", s, e, a.z)
		}
	}
}

func TestBigIntSetZ(t *testing.T) {
	for _, a := range sumZZ {
		var z BigInt
		z.Set(a.z)
		if (&z).Cmp(a.z) != 0 {
			t.Errorf("got z = %v; want %v", &z, a.z)
		}
	}
}

func TestBigIntAbsZ(t *testing.T) {
	var zero BigInt
	for _, a := range sumZZ {
		var z BigInt
		z.Abs(a.z)
		var e BigInt
		e.Set(a.z)
		if e.Cmp(&zero) < 0 {
			e.Sub(&zero, &e)
		}
		if z.Cmp(&e) != 0 {
			t.Errorf("got z = %v; want %v", &z, &e)
		}
	}
}

func testFunZZ(t *testing.T, msg string, f funZZ, a argZZ) {
	var z BigInt
	f(&z, a.x, a.y)
	if (&z).Cmp(a.z) != 0 {
		t.Errorf("%s%+v\n\tgot z = %v; want %v", msg, a, &z, a.z)
	}
}

func TestBigIntSumZZ(t *testing.T) {
	AddZZ := func(z, x, y *BigInt) *BigInt { return z.Add(x, y) }
	SubZZ := func(z, x, y *BigInt) *BigInt { return z.Sub(x, y) }
	for _, a := range sumZZ {
		arg := a
		testFunZZ(t, "AddZZ", AddZZ, arg)

		arg = argZZ{a.z, a.y, a.x}
		testFunZZ(t, "AddZZ symmetric", AddZZ, arg)

		arg = argZZ{a.x, a.z, a.y}
		testFunZZ(t, "SubZZ", SubZZ, arg)

		arg = argZZ{a.y, a.z, a.x}
		testFunZZ(t, "SubZZ symmetric", SubZZ, arg)
	}
}

func TestBigIntProdZZ(t *testing.T) {
	MulZZ := func(z, x, y *BigInt) *BigInt { return z.Mul(x, y) }
	for _, a := range prodZZ {
		arg := a
		testFunZZ(t, "MulZZ", MulZZ, arg)

		arg = argZZ{a.z, a.y, a.x}
		testFunZZ(t, "MulZZ symmetric", MulZZ, arg)
	}
}

// mulBytes returns x*y via grade school multiplication. Both inputs
// and the result are assumed to be in big-endian representation (to
// match the semantics of BigInt.Bytes and BigInt.SetBytes).
func mulBytes(x, y []byte) []byte {
	z := make([]byte, len(x)+len(y))

	// multiply
	k0 := len(z) - 1
	for j := len(y) - 1; j >= 0; j-- {
		d := int(y[j])
		if d != 0 {
			k := k0
			carry := 0
			for i := len(x) - 1; i >= 0; i-- {
				t := int(z[k]) + int(x[i])*d + carry
				z[k], carry = byte(t), t>>8
				k--
			}
			z[k] = byte(carry)
		}
		k0--
	}

	// normalize (remove leading 0's)
	i := 0
	for i < len(z) && z[i] == 0 {
		i++
	}

	return z[i:]
}

func checkMul(a, b []byte) bool {
	var x, y, z1 BigInt
	x.SetBytes(a)
	y.SetBytes(b)
	z1.Mul(&x, &y)

	var z2 BigInt
	z2.SetBytes(mulBytes(a, b))

	return z1.Cmp(&z2) == 0
}

func TestBigIntMul(t *testing.T) {
	if err := quick.Check(checkMul, nil); err != nil {
		t.Error(err)
	}
}

var mulRangesN = []struct {
	a, b uint64
	prod string
}{
	{0, 0, "0"},
	{1, 1, "1"},
	{1, 2, "2"},
	{1, 3, "6"},
	{10, 10, "10"},
	{0, 100, "0"},
	{0, 1e9, "0"},
	{1, 0, "1"},                    // empty range
	{100, 1, "1"},                  // empty range
	{1, 10, "3628800"},             // 10!
	{1, 20, "2432902008176640000"}, // 20!
	{1, 100,
		"933262154439441526816992388562667004907159682643816214685929" +
			"638952175999932299156089414639761565182862536979208272237582" +
			"51185210916864000000000000000000000000", // 100!
	},
}

var mulRangesZ = []struct {
	a, b int64
	prod string
}{
	// entirely positive ranges are covered by mulRangesN
	{-1, 1, "0"},
	{-2, -1, "2"},
	{-3, -2, "6"},
	{-3, -1, "-6"},
	{1, 3, "6"},
	{-10, -10, "-10"},
	{0, -1, "1"},                      // empty range
	{-1, -100, "1"},                   // empty range
	{-1, 1, "0"},                      // range includes 0
	{-1e9, 0, "0"},                    // range includes 0
	{-1e9, 1e9, "0"},                  // range includes 0
	{-10, -1, "3628800"},              // 10!
	{-20, -2, "-2432902008176640000"}, // -20!
	{-99, -1,
		"-933262154439441526816992388562667004907159682643816214685929" +
			"638952175999932299156089414639761565182862536979208272237582" +
			"511852109168640000000000000000000000", // -99!
	},
}

func TestBigIntMulRangeZ(t *testing.T) {
	var tmp BigInt
	// test entirely positive ranges
	for i, r := range mulRangesN {
		prod := tmp.MulRange(int64(r.a), int64(r.b)).String()
		if prod != r.prod {
			t.Errorf("#%da: got %s; want %s", i, prod, r.prod)
		}
	}
	// test other ranges
	for i, r := range mulRangesZ {
		prod := tmp.MulRange(r.a, r.b).String()
		if prod != r.prod {
			t.Errorf("#%db: got %s; want %s", i, prod, r.prod)
		}
	}
}

func TestBigIntBinomial(t *testing.T) {
	var z BigInt
	for _, test := range []struct {
		n, k int64
		want string
	}{
		{0, 0, "1"},
		{0, 1, "0"},
		{1, 0, "1"},
		{1, 1, "1"},
		{1, 10, "0"},
		{4, 0, "1"},
		{4, 1, "4"},
		{4, 2, "6"},
		{4, 3, "4"},
		{4, 4, "1"},
		{10, 1, "10"},
		{10, 9, "10"},
		{10, 5, "252"},
		{11, 5, "462"},
		{11, 6, "462"},
		{100, 10, "17310309456440"},
		{100, 90, "17310309456440"},
		{1000, 10, "263409560461970212832400"},
		{1000, 990, "263409560461970212832400"},
	} {
		if got := z.Binomial(test.n, test.k).String(); got != test.want {
			t.Errorf("Binomial(%d, %d) = %s; want %s", test.n, test.k, got, test.want)
		}
	}
}

// Examples from the Go Language Spec, section "Arithmetic operators"
var divisionSignsTests = []struct {
	x, y int64
	q, r int64 // T-division
	d, m int64 // Euclidean division
}{
	{5, 3, 1, 2, 1, 2},
	{-5, 3, -1, -2, -2, 1},
	{5, -3, -1, 2, -1, 2},
	{-5, -3, 1, -2, 2, 1},
	{1, 2, 0, 1, 0, 1},
	{8, 4, 2, 0, 2, 0},
}

func TestBigIntDivisionSigns(t *testing.T) {
	for i, test := range divisionSignsTests {
		x := NewBigInt(test.x)
		y := NewBigInt(test.y)
		q := NewBigInt(test.q)
		r := NewBigInt(test.r)
		d := NewBigInt(test.d)
		m := NewBigInt(test.m)

		q1 := new(BigInt).Quo(x, y)
		r1 := new(BigInt).Rem(x, y)
		if q1.Cmp(q) != 0 || r1.Cmp(r) != 0 {
			t.Errorf("#%d QuoRem: got (%s, %s), want (%s, %s)", i, q1, r1, q, r)
		}

		q2, r2 := new(BigInt).QuoRem(x, y, new(BigInt))
		if q2.Cmp(q) != 0 || r2.Cmp(r) != 0 {
			t.Errorf("#%d QuoRem: got (%s, %s), want (%s, %s)", i, q2, r2, q, r)
		}

		d1 := new(BigInt).Div(x, y)
		m1 := new(BigInt).Mod(x, y)
		if d1.Cmp(d) != 0 || m1.Cmp(m) != 0 {
			t.Errorf("#%d DivMod: got (%s, %s), want (%s, %s)", i, d1, m1, d, m)
		}

		d2, m2 := new(BigInt).DivMod(x, y, new(BigInt))
		if d2.Cmp(d) != 0 || m2.Cmp(m) != 0 {
			t.Errorf("#%d DivMod: got (%s, %s), want (%s, %s)", i, d2, m2, d, m)
		}
	}
}

func checkSetBytes(b []byte) bool {
	hex1 := hex.EncodeToString(new(BigInt).SetBytes(b).Bytes())
	hex2 := hex.EncodeToString(b)

	for len(hex1) < len(hex2) {
		hex1 = "0" + hex1
	}

	for len(hex1) > len(hex2) {
		hex2 = "0" + hex2
	}

	return hex1 == hex2
}

func TestBigIntSetBytes(t *testing.T) {
	if err := quick.Check(checkSetBytes, nil); err != nil {
		t.Error(err)
	}
}

func checkBytes(b []byte) bool {
	// trim leading zero bytes since Bytes() won't return them
	// (was issue 12231)
	for len(b) > 0 && b[0] == 0 {
		b = b[1:]
	}
	b2 := new(BigInt).SetBytes(b).Bytes()
	return bytes.Equal(b, b2)
}

func TestBigIntBytes(t *testing.T) {
	if err := quick.Check(checkBytes, nil); err != nil {
		t.Error(err)
	}
}

func checkQuo(x, y []byte) bool {
	u := new(BigInt).SetBytes(x)
	v := new(BigInt).SetBytes(y)

	var tmp1 big.Int
	if len(v.inner(&tmp1).Bits()) == 0 {
		return true
	}

	r := new(BigInt)
	q, r := new(BigInt).QuoRem(u, v, r)

	if r.Cmp(v) >= 0 {
		return false
	}

	uprime := new(BigInt).Set(q)
	uprime.Mul(uprime, v)
	uprime.Add(uprime, r)

	return uprime.Cmp(u) == 0
}

var quoTests = []struct {
	x, y string
	q, r string
}{
	{
		"476217953993950760840509444250624797097991362735329973741718102894495832294430498335824897858659711275234906400899559094370964723884706254265559534144986498357",
		"9353930466774385905609975137998169297361893554149986716853295022578535724979483772383667534691121982974895531435241089241440253066816724367338287092081996",
		"50911",
		"1",
	},
	{
		"11510768301994997771168",
		"1328165573307167369775",
		"8",
		"885443715537658812968",
	},
}

func TestBigIntQuo(t *testing.T) {
	if err := quick.Check(checkQuo, nil); err != nil {
		t.Error(err)
	}

	for i, test := range quoTests {
		x, _ := new(BigInt).SetString(test.x, 10)
		y, _ := new(BigInt).SetString(test.y, 10)
		expectedQ, _ := new(BigInt).SetString(test.q, 10)
		expectedR, _ := new(BigInt).SetString(test.r, 10)

		r := new(BigInt)
		q, r := new(BigInt).QuoRem(x, y, r)

		if q.Cmp(expectedQ) != 0 || r.Cmp(expectedR) != 0 {
			t.Errorf("#%d got (%s, %s) want (%s, %s)", i, q, r, expectedQ, expectedR)
		}
	}
}

var bitLenTests = []struct {
	in  string
	out int
}{
	{"-1", 1},
	{"0", 0},
	{"1", 1},
	{"2", 2},
	{"4", 3},
	{"0xabc", 12},
	{"0x8000", 16},
	{"0x80000000", 32},
	{"0x800000000000", 48},
	{"0x8000000000000000", 64},
	{"0x80000000000000000000", 80},
	{"-0x4000000000000000000000", 87},
}

func TestBigIntBitLen(t *testing.T) {
	for i, test := range bitLenTests {
		x, ok := new(BigInt).SetString(test.in, 0)
		if !ok {
			t.Errorf("#%d test input invalid: %s", i, test.in)
			continue
		}

		if n := x.BitLen(); n != test.out {
			t.Errorf("#%d got %d want %d", i, n, test.out)
		}
	}
}

var expTests = []struct {
	x, y, m string
	out     string
}{
	// y <= 0
	{"0", "0", "", "1"},
	{"1", "0", "", "1"},
	{"-10", "0", "", "1"},
	{"1234", "-1", "", "1"},
	{"1234", "-1", "0", "1"},
	{"17", "-100", "1234", "865"},
	{"2", "-100", "1234", ""},

	// m == 1
	{"0", "0", "1", "0"},
	{"1", "0", "1", "0"},
	{"-10", "0", "1", "0"},
	{"1234", "-1", "1", "0"},

	// misc
	{"5", "1", "3", "2"},
	{"5", "-7", "", "1"},
	{"-5", "-7", "", "1"},
	{"5", "0", "", "1"},
	{"-5", "0", "", "1"},
	{"5", "1", "", "5"},
	{"-5", "1", "", "-5"},
	{"-5", "1", "7", "2"},
	{"-2", "3", "2", "0"},
	{"5", "2", "", "25"},
	{"1", "65537", "2", "1"},
	{"0x8000000000000000", "2", "", "0x40000000000000000000000000000000"},
	{"0x8000000000000000", "2", "6719", "4944"},
	{"0x8000000000000000", "3", "6719", "5447"},
	{"0x8000000000000000", "1000", "6719", "1603"},
	{"0x8000000000000000", "1000000", "6719", "3199"},
	{"0x8000000000000000", "-1000000", "6719", "3663"}, // 3663 = ModInverse(3199, 6719) Issue #25865

	{"0xffffffffffffffffffffffffffffffff", "0x12345678123456781234567812345678123456789", "0x01112222333344445555666677778889", "0x36168FA1DB3AAE6C8CE647E137F97A"},

	{
		"2938462938472983472983659726349017249287491026512746239764525612965293865296239471239874193284792387498274256129746192347",
		"298472983472983471903246121093472394872319615612417471234712061",
		"29834729834729834729347290846729561262544958723956495615629569234729836259263598127342374289365912465901365498236492183464",
		"23537740700184054162508175125554701713153216681790245129157191391322321508055833908509185839069455749219131480588829346291",
	},
	// test case for issue 8822
	{
		"11001289118363089646017359372117963499250546375269047542777928006103246876688756735760905680604646624353196869572752623285140408755420374049317646428185270079555372763503115646054602867593662923894140940837479507194934267532831694565516466765025434902348314525627418515646588160955862839022051353653052947073136084780742729727874803457643848197499548297570026926927502505634297079527299004267769780768565695459945235586892627059178884998772989397505061206395455591503771677500931269477503508150175717121828518985901959919560700853226255420793148986854391552859459511723547532575574664944815966793196961286234040892865",
		"0xB08FFB20760FFED58FADA86DFEF71AD72AA0FA763219618FE022C197E54708BB1191C66470250FCE8879487507CEE41381CA4D932F81C2B3F1AB20B539D50DCD",
		"0xAC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF73",
		"21484252197776302499639938883777710321993113097987201050501182909581359357618579566746556372589385361683610524730509041328855066514963385522570894839035884713051640171474186548713546686476761306436434146475140156284389181808675016576845833340494848283681088886584219750554408060556769486628029028720727393293111678826356480455433909233520504112074401376133077150471237549474149190242010469539006449596611576612573955754349042329130631128234637924786466585703488460540228477440853493392086251021228087076124706778899179648655221663765993962724699135217212118535057766739392069738618682722216712319320435674779146070442",
	},
	{
		"-0x1BCE04427D8032319A89E5C4136456671AC620883F2C4139E57F91307C485AD2D6204F4F87A58262652DB5DBBAC72B0613E51B835E7153BEC6068F5C8D696B74DBD18FEC316AEF73985CF0475663208EB46B4F17DD9DA55367B03323E5491A70997B90C059FB34809E6EE55BCFBD5F2F52233BFE62E6AA9E4E26A1D4C2439883D14F2633D55D8AA66A1ACD5595E778AC3A280517F1157989E70C1A437B849F1877B779CC3CDDEDE2DAA6594A6C66D181A00A5F777EE60596D8773998F6E988DEAE4CCA60E4DDCF9590543C89F74F603259FCAD71660D30294FBBE6490300F78A9D63FA660DC9417B8B9DDA28BEB3977B621B988E23D4D954F322C3540541BC649ABD504C50FADFD9F0987D58A2BF689313A285E773FF02899A6EF887D1D4A0D2",
		"0xB08FFB20760FFED58FADA86DFEF71AD72AA0FA763219618FE022C197E54708BB1191C66470250FCE8879487507CEE41381CA4D932F81C2B3F1AB20B539D50DCD",
		"0xAC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF73",
		"21484252197776302499639938883777710321993113097987201050501182909581359357618579566746556372589385361683610524730509041328855066514963385522570894839035884713051640171474186548713546686476761306436434146475140156284389181808675016576845833340494848283681088886584219750554408060556769486628029028720727393293111678826356480455433909233520504112074401376133077150471237549474149190242010469539006449596611576612573955754349042329130631128234637924786466585703488460540228477440853493392086251021228087076124706778899179648655221663765993962724699135217212118535057766739392069738618682722216712319320435674779146070442",
	},

	// test cases for issue 13907
	{"0xffffffff00000001", "0xffffffff00000001", "0xffffffff00000001", "0"},
	{"0xffffffffffffffff00000001", "0xffffffffffffffff00000001", "0xffffffffffffffff00000001", "0"},
	{"0xffffffffffffffffffffffff00000001", "0xffffffffffffffffffffffff00000001", "0xffffffffffffffffffffffff00000001", "0"},
	{"0xffffffffffffffffffffffffffffffff00000001", "0xffffffffffffffffffffffffffffffff00000001", "0xffffffffffffffffffffffffffffffff00000001", "0"},

	{
		"2",
		"0xB08FFB20760FFED58FADA86DFEF71AD72AA0FA763219618FE022C197E54708BB1191C66470250FCE8879487507CEE41381CA4D932F81C2B3F1AB20B539D50DCD",
		"0xAC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF73", // odd
		"0x6AADD3E3E424D5B713FCAA8D8945B1E055166132038C57BBD2D51C833F0C5EA2007A2324CE514F8E8C2F008A2F36F44005A4039CB55830986F734C93DAF0EB4BAB54A6A8C7081864F44346E9BC6F0A3EB9F2C0146A00C6A05187D0C101E1F2D038CDB70CB5E9E05A2D188AB6CBB46286624D4415E7D4DBFAD3BCC6009D915C406EED38F468B940F41E6BEDC0430DD78E6F19A7DA3A27498A4181E24D738B0072D8F6ADB8C9809A5B033A09785814FD9919F6EF9F83EEA519BEC593855C4C10CBEEC582D4AE0792158823B0275E6AEC35242740468FAF3D5C60FD1E376362B6322F78B7ED0CA1C5BBCD2B49734A56C0967A1D01A100932C837B91D592CE08ABFF",
	},
	{
		"2",
		"0xB08FFB20760FFED58FADA86DFEF71AD72AA0FA763219618FE022C197E54708BB1191C66470250FCE8879487507CEE41381CA4D932F81C2B3F1AB20B539D50DCD",
		"0xAC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF72", // even
		"0x7858794B5897C29F4ED0B40913416AB6C48588484E6A45F2ED3E26C941D878E923575AAC434EE2750E6439A6976F9BB4D64CEDB2A53CE8D04DD48CADCDF8E46F22747C6B81C6CEA86C0D873FBF7CEF262BAAC43A522BD7F32F3CDAC52B9337C77B3DCFB3DB3EDD80476331E82F4B1DF8EFDC1220C92656DFC9197BDC1877804E28D928A2A284B8DED506CBA304435C9D0133C246C98A7D890D1DE60CBC53A024361DA83A9B8775019083D22AC6820ED7C3C68F8E801DD4EC779EE0A05C6EB682EF9840D285B838369BA7E148FA27691D524FAEAF7C6ECE2A4B99A294B9F2C241857B5B90CC8BFFCFCF18DFA7D676131D5CD3855A5A3E8EBFA0CDFADB4D198B4A",
	},
}

func TestBigIntExp(t *testing.T) {
	for i, test := range expTests {
		x, ok1 := new(BigInt).SetString(test.x, 0)
		y, ok2 := new(BigInt).SetString(test.y, 0)

		var ok3, ok4 bool
		var out, m *BigInt

		if len(test.out) == 0 {
			out, ok3 = nil, true
		} else {
			out, ok3 = new(BigInt).SetString(test.out, 0)
		}

		if len(test.m) == 0 {
			m, ok4 = nil, true
		} else {
			m, ok4 = new(BigInt).SetString(test.m, 0)
		}

		if !ok1 || !ok2 || !ok3 || !ok4 {
			t.Errorf("#%d: error in input", i)
			continue
		}

		z1 := new(BigInt).Exp(x, y, m)
		if !(z1 == nil && out == nil || z1.Cmp(out) == 0) {
			t.Errorf("#%d: got %x want %x", i, z1, out)
		}

		if m == nil {
			// The result should be the same as for m == 0;
			// specifically, there should be no div-zero panic.
			m = new(BigInt) // m != nil && len(m.abs) == 0
			z2 := new(BigInt).Exp(x, y, m)
			if z2.Cmp(z1) != 0 {
				t.Errorf("#%d: got %x want %x", i, z2, z1)
			}
		}
	}
}

type intShiftTest struct {
	in    string
	shift uint
	out   string
}

var rshTests = []intShiftTest{
	{"0", 0, "0"},
	{"-0", 0, "0"},
	{"0", 1, "0"},
	{"0", 2, "0"},
	{"1", 0, "1"},
	{"1", 1, "0"},
	{"1", 2, "0"},
	{"2", 0, "2"},
	{"2", 1, "1"},
	{"-1", 0, "-1"},
	{"-1", 1, "-1"},
	{"-1", 10, "-1"},
	{"-100", 2, "-25"},
	{"-100", 3, "-13"},
	{"-100", 100, "-1"},
	{"4294967296", 0, "4294967296"},
	{"4294967296", 1, "2147483648"},
	{"4294967296", 2, "1073741824"},
	{"18446744073709551616", 0, "18446744073709551616"},
	{"18446744073709551616", 1, "9223372036854775808"},
	{"18446744073709551616", 2, "4611686018427387904"},
	{"18446744073709551616", 64, "1"},
	{"340282366920938463463374607431768211456", 64, "18446744073709551616"},
	{"340282366920938463463374607431768211456", 128, "1"},
}

func TestBigIntRsh(t *testing.T) {
	for i, test := range rshTests {
		in, _ := new(BigInt).SetString(test.in, 10)
		expected, _ := new(BigInt).SetString(test.out, 10)
		out := new(BigInt).Rsh(in, test.shift)

		if out.Cmp(expected) != 0 {
			t.Errorf("#%d: got %s want %s", i, out, expected)
		}
	}
}

func TestBigIntRshSelf(t *testing.T) {
	for i, test := range rshTests {
		z, _ := new(BigInt).SetString(test.in, 10)
		expected, _ := new(BigInt).SetString(test.out, 10)
		z.Rsh(z, test.shift)

		if z.Cmp(expected) != 0 {
			t.Errorf("#%d: got %s want %s", i, z, expected)
		}
	}
}

var lshTests = []intShiftTest{
	{"0", 0, "0"},
	{"0", 1, "0"},
	{"0", 2, "0"},
	{"1", 0, "1"},
	{"1", 1, "2"},
	{"1", 2, "4"},
	{"2", 0, "2"},
	{"2", 1, "4"},
	{"2", 2, "8"},
	{"-87", 1, "-174"},
	{"4294967296", 0, "4294967296"},
	{"4294967296", 1, "8589934592"},
	{"4294967296", 2, "17179869184"},
	{"18446744073709551616", 0, "18446744073709551616"},
	{"9223372036854775808", 1, "18446744073709551616"},
	{"4611686018427387904", 2, "18446744073709551616"},
	{"1", 64, "18446744073709551616"},
	{"18446744073709551616", 64, "340282366920938463463374607431768211456"},
	{"1", 128, "340282366920938463463374607431768211456"},
}

func TestBigIntLsh(t *testing.T) {
	for i, test := range lshTests {
		in, _ := new(BigInt).SetString(test.in, 10)
		expected, _ := new(BigInt).SetString(test.out, 10)
		out := new(BigInt).Lsh(in, test.shift)

		if out.Cmp(expected) != 0 {
			t.Errorf("#%d: got %s want %s", i, out, expected)
		}
	}
}

func TestBigIntLshSelf(t *testing.T) {
	for i, test := range lshTests {
		z, _ := new(BigInt).SetString(test.in, 10)
		expected, _ := new(BigInt).SetString(test.out, 10)
		z.Lsh(z, test.shift)

		if z.Cmp(expected) != 0 {
			t.Errorf("#%d: got %s want %s", i, z, expected)
		}
	}
}

func TestBigIntLshRsh(t *testing.T) {
	for i, test := range rshTests {
		in, _ := new(BigInt).SetString(test.in, 10)
		out := new(BigInt).Lsh(in, test.shift)
		out = out.Rsh(out, test.shift)

		if in.Cmp(out) != 0 {
			t.Errorf("#%d: got %s want %s", i, out, in)
		}
	}
	for i, test := range lshTests {
		in, _ := new(BigInt).SetString(test.in, 10)
		out := new(BigInt).Lsh(in, test.shift)
		out.Rsh(out, test.shift)

		if in.Cmp(out) != 0 {
			t.Errorf("#%d: got %s want %s", i, out, in)
		}
	}
}

// Entries must be sorted by value in ascending order.
var cmpAbsTests = []string{
	"0",
	"1",
	"2",
	"10",
	"10000000",
	"2783678367462374683678456387645876387564783686583485",
	"2783678367462374683678456387645876387564783686583486",
	"32957394867987420967976567076075976570670947609750670956097509670576075067076027578341538",
}

func TestBigIntCmpAbs(t *testing.T) {
	values := make([]*BigInt, len(cmpAbsTests))
	var prev *BigInt
	for i, s := range cmpAbsTests {
		x, ok := new(BigInt).SetString(s, 0)
		if !ok {
			t.Fatalf("SetString(%s, 0) failed", s)
		}
		if prev != nil && prev.Cmp(x) >= 0 {
			t.Fatal("cmpAbsTests entries not sorted in ascending order")
		}
		values[i] = x
		prev = x
	}

	for i, x := range values {
		for j, y := range values {
			// try all combinations of signs for x, y
			for k := 0; k < 4; k++ {
				var a, b BigInt
				a.Set(x)
				b.Set(y)
				if k&1 != 0 {
					a.Neg(&a)
				}
				if k&2 != 0 {
					b.Neg(&b)
				}

				got := a.CmpAbs(&b)
				want := 0
				switch {
				case i > j:
					want = 1
				case i < j:
					want = -1
				}
				if got != want {
					t.Errorf("absCmp |%s|, |%s|: got %d; want %d", &a, &b, got, want)
				}
			}
		}
	}
}

func TestBigIntCmpSelf(t *testing.T) {
	for _, s := range cmpAbsTests {
		x, ok := new(BigInt).SetString(s, 0)
		if !ok {
			t.Fatalf("SetString(%s, 0) failed", s)
		}
		got := x.Cmp(x)
		want := 0
		if got != want {
			t.Errorf("x = %s: x.Cmp(x): got %d; want %d", x, got, want)
		}
	}
}

var int64Tests = []string{
	// int64
	"0",
	"1",
	"-1",
	"4294967295",
	"-4294967295",
	"4294967296",
	"-4294967296",
	"9223372036854775807",
	"-9223372036854775807",
	"-9223372036854775808",

	// not int64
	"0x8000000000000000",
	"-0x8000000000000001",
	"38579843757496759476987459679745",
	"-38579843757496759476987459679745",
}

func TestBigInt64(t *testing.T) {
	for _, s := range int64Tests {
		var x BigInt
		_, ok := x.SetString(s, 0)
		if !ok {
			t.Errorf("SetString(%s, 0) failed", s)
			continue
		}

		want, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrRange {
				if x.IsInt64() {
					t.Errorf("IsInt64(%s) succeeded unexpectedly", s)
				}
			} else {
				t.Errorf("ParseInt(%s) failed", s)
			}
			continue
		}

		if !x.IsInt64() {
			t.Errorf("IsInt64(%s) failed unexpectedly", s)
		}

		got := x.Int64()
		if got != want {
			t.Errorf("Int64(%s) = %d; want %d", s, got, want)
		}
	}
}

var uint64Tests = []string{
	// uint64
	"0",
	"1",
	"4294967295",
	"4294967296",
	"8589934591",
	"8589934592",
	"9223372036854775807",
	"9223372036854775808",
	"0x08000000000000000",

	// not uint64
	"0x10000000000000000",
	"-0x08000000000000000",
	"-1",
}

func TestBigIntUint64(t *testing.T) {
	for _, s := range uint64Tests {
		var x BigInt
		_, ok := x.SetString(s, 0)
		if !ok {
			t.Errorf("SetString(%s, 0) failed", s)
			continue
		}

		want, err := strconv.ParseUint(s, 0, 64)
		if err != nil {
			// check for sign explicitly (ErrRange doesn't cover signed input)
			if s[0] == '-' || err.(*strconv.NumError).Err == strconv.ErrRange {
				if x.IsUint64() {
					t.Errorf("IsUint64(%s) succeeded unexpectedly", s)
				}
			} else {
				t.Errorf("ParseUint(%s) failed", s)
			}
			continue
		}

		if !x.IsUint64() {
			t.Errorf("IsUint64(%s) failed unexpectedly", s)
		}

		got := x.Uint64()
		if got != want {
			t.Errorf("Uint64(%s) = %d; want %d", s, got, want)
		}
	}
}

var bitwiseTests = []struct {
	x, y                 string
	and, or, xor, andNot string
}{
	{"0x00", "0x00", "0x00", "0x00", "0x00", "0x00"},
	{"0x00", "0x01", "0x00", "0x01", "0x01", "0x00"},
	{"0x01", "0x00", "0x00", "0x01", "0x01", "0x01"},
	{"-0x01", "0x00", "0x00", "-0x01", "-0x01", "-0x01"},
	{"-0xaf", "-0x50", "-0xf0", "-0x0f", "0xe1", "0x41"},
	{"0x00", "-0x01", "0x00", "-0x01", "-0x01", "0x00"},
	{"0x01", "0x01", "0x01", "0x01", "0x00", "0x00"},
	{"-0x01", "-0x01", "-0x01", "-0x01", "0x00", "0x00"},
	{"0x07", "0x08", "0x00", "0x0f", "0x0f", "0x07"},
	{"0x05", "0x0f", "0x05", "0x0f", "0x0a", "0x00"},
	{"0xff", "-0x0a", "0xf6", "-0x01", "-0xf7", "0x09"},
	{"0x013ff6", "0x9a4e", "0x1a46", "0x01bffe", "0x01a5b8", "0x0125b0"},
	{"-0x013ff6", "0x9a4e", "0x800a", "-0x0125b2", "-0x01a5bc", "-0x01c000"},
	{"-0x013ff6", "-0x9a4e", "-0x01bffe", "-0x1a46", "0x01a5b8", "0x8008"},
	{
		"0x1000009dc6e3d9822cba04129bcbe3401",
		"0xb9bd7d543685789d57cb918e833af352559021483cdb05cc21fd",
		"0x1000001186210100001000009048c2001",
		"0xb9bd7d543685789d57cb918e8bfeff7fddb2ebe87dfbbdfe35fd",
		"0xb9bd7d543685789d57ca918e8ae69d6fcdb2eae87df2b97215fc",
		"0x8c40c2d8822caa04120b8321400",
	},
	{
		"0x1000009dc6e3d9822cba04129bcbe3401",
		"-0xb9bd7d543685789d57cb918e833af352559021483cdb05cc21fd",
		"0x8c40c2d8822caa04120b8321401",
		"-0xb9bd7d543685789d57ca918e82229142459020483cd2014001fd",
		"-0xb9bd7d543685789d57ca918e8ae69d6fcdb2eae87df2b97215fe",
		"0x1000001186210100001000009048c2000",
	},
	{
		"-0x1000009dc6e3d9822cba04129bcbe3401",
		"-0xb9bd7d543685789d57cb918e833af352559021483cdb05cc21fd",
		"-0xb9bd7d543685789d57cb918e8bfeff7fddb2ebe87dfbbdfe35fd",
		"-0x1000001186210100001000009048c2001",
		"0xb9bd7d543685789d57ca918e8ae69d6fcdb2eae87df2b97215fc",
		"0xb9bd7d543685789d57ca918e82229142459020483cd2014001fc",
	},
}

type bitFun func(z, x, y *BigInt) *BigInt

func testBitFun(t *testing.T, msg string, f bitFun, x, y *BigInt, exp string) {
	expected := new(BigInt)
	expected.SetString(exp, 0)

	out := f(new(BigInt), x, y)
	if out.Cmp(expected) != 0 {
		t.Errorf("%s: got %s want %s", msg, out, expected)
	}
}

func testBitFunSelf(t *testing.T, msg string, f bitFun, x, y *BigInt, exp string) {
	self := new(BigInt)
	self.Set(x)
	expected := new(BigInt)
	expected.SetString(exp, 0)

	self = f(self, self, y)
	if self.Cmp(expected) != 0 {
		t.Errorf("%s: got %s want %s", msg, self, expected)
	}
}

func altBit(x *BigInt, i int) uint {
	z := new(BigInt).Rsh(x, uint(i))
	z = z.And(z, NewBigInt(1))
	if z.Cmp(new(BigInt)) != 0 {
		return 1
	}
	return 0
}

func altSetBit(z *BigInt, x *BigInt, i int, b uint) *BigInt {
	one := NewBigInt(1)
	m := one.Lsh(one, uint(i))
	switch b {
	case 1:
		return z.Or(x, m)
	case 0:
		return z.AndNot(x, m)
	}
	panic("set bit is not 0 or 1")
}

func testBitset(t *testing.T, x *BigInt) {
	n := x.BitLen()
	z := new(BigInt).Set(x)
	z1 := new(BigInt).Set(x)
	for i := 0; i < n+10; i++ {
		old := z.Bit(i)
		old1 := altBit(z1, i)
		if old != old1 {
			t.Errorf("bitset: inconsistent value for Bit(%s, %d), got %v want %v", z1, i, old, old1)
		}
		z := new(BigInt).SetBit(z, i, 1)
		z1 := altSetBit(new(BigInt), z1, i, 1)
		if z.Bit(i) == 0 {
			t.Errorf("bitset: bit %d of %s got 0 want 1", i, x)
		}
		if z.Cmp(z1) != 0 {
			t.Errorf("bitset: inconsistent value after SetBit 1, got %s want %s", z, z1)
		}
		z.SetBit(z, i, 0)
		altSetBit(z1, z1, i, 0)
		if z.Bit(i) != 0 {
			t.Errorf("bitset: bit %d of %s got 1 want 0", i, x)
		}
		if z.Cmp(z1) != 0 {
			t.Errorf("bitset: inconsistent value after SetBit 0, got %s want %s", z, z1)
		}
		altSetBit(z1, z1, i, old)
		z.SetBit(z, i, old)
		if z.Cmp(z1) != 0 {
			t.Errorf("bitset: inconsistent value after SetBit old, got %s want %s", z, z1)
		}
	}
	if z.Cmp(x) != 0 {
		t.Errorf("bitset: got %s want %s", z, x)
	}
}

var bitsetTests = []struct {
	x string
	i int
	b uint
}{
	{"0", 0, 0},
	{"0", 200, 0},
	{"1", 0, 1},
	{"1", 1, 0},
	{"-1", 0, 1},
	{"-1", 200, 1},
	{"0x2000000000000000000000000000", 108, 0},
	{"0x2000000000000000000000000000", 109, 1},
	{"0x2000000000000000000000000000", 110, 0},
	{"-0x2000000000000000000000000001", 108, 1},
	{"-0x2000000000000000000000000001", 109, 0},
	{"-0x2000000000000000000000000001", 110, 1},
}

func TestBigIntBitSet(t *testing.T) {
	for _, test := range bitwiseTests {
		x := new(BigInt)
		x.SetString(test.x, 0)
		testBitset(t, x)
		x = new(BigInt)
		x.SetString(test.y, 0)
		testBitset(t, x)
	}
	for i, test := range bitsetTests {
		x := new(BigInt)
		x.SetString(test.x, 0)
		b := x.Bit(test.i)
		if b != test.b {
			t.Errorf("#%d got %v want %v", i, b, test.b)
		}
	}
	z := NewBigInt(1)
	z.SetBit(NewBigInt(0), 2, 1)
	if z.Cmp(NewBigInt(4)) != 0 {
		t.Errorf("destination leaked into result; got %s want 4", z)
	}
}

var tzbTests = []struct {
	in  string
	out uint
}{
	{"0", 0},
	{"1", 0},
	{"-1", 0},
	{"4", 2},
	{"-8", 3},
	{"0x4000000000000000000", 74},
	{"-0x8000000000000000000", 75},
}

func TestBigIntTrailingZeroBits(t *testing.T) {
	for i, test := range tzbTests {
		in, _ := new(BigInt).SetString(test.in, 0)
		want := test.out
		got := in.TrailingZeroBits()

		if got != want {
			t.Errorf("#%d: got %v want %v", i, got, want)
		}
	}
}

func TestBigIntBitwise(t *testing.T) {
	x := new(BigInt)
	y := new(BigInt)
	for _, test := range bitwiseTests {
		x.SetString(test.x, 0)
		y.SetString(test.y, 0)

		testBitFun(t, "and", (*BigInt).And, x, y, test.and)
		testBitFunSelf(t, "and", (*BigInt).And, x, y, test.and)
		testBitFun(t, "andNot", (*BigInt).AndNot, x, y, test.andNot)
		testBitFunSelf(t, "andNot", (*BigInt).AndNot, x, y, test.andNot)
		testBitFun(t, "or", (*BigInt).Or, x, y, test.or)
		testBitFunSelf(t, "or", (*BigInt).Or, x, y, test.or)
		testBitFun(t, "xor", (*BigInt).Xor, x, y, test.xor)
		testBitFunSelf(t, "xor", (*BigInt).Xor, x, y, test.xor)
	}
}

var notTests = []struct {
	in  string
	out string
}{
	{"0", "-1"},
	{"1", "-2"},
	{"7", "-8"},
	{"0", "-1"},
	{"-81910", "81909"},
	{
		"298472983472983471903246121093472394872319615612417471234712061",
		"-298472983472983471903246121093472394872319615612417471234712062",
	},
}

func TestBigIntNot(t *testing.T) {
	in := new(BigInt)
	out := new(BigInt)
	expected := new(BigInt)
	for i, test := range notTests {
		in.SetString(test.in, 10)
		expected.SetString(test.out, 10)
		out = out.Not(in)
		if out.Cmp(expected) != 0 {
			t.Errorf("#%d: got %s want %s", i, out, expected)
		}
		out = out.Not(out)
		if out.Cmp(in) != 0 {
			t.Errorf("#%d: got %s want %s", i, out, in)
		}
	}
}

var modInverseTests = []struct {
	element string
	modulus string
}{
	{"1234567", "458948883992"},
	{"239487239847", "2410312426921032588552076022197566074856950548502459942654116941958108831682612228890093858261341614673227141477904012196503648957050582631942730706805009223062734745341073406696246014589361659774041027169249453200378729434170325843778659198143763193776859869524088940195577346119843545301547043747207749969763750084308926339295559968882457872412993810129130294592999947926365264059284647209730384947211681434464714438488520940127459844288859336526896320919633919"},
	{"-10", "13"}, // issue #16984
	{"10", "-13"},
	{"-17", "-13"},
}

func TestBigIntModInverse(t *testing.T) {
	var element, modulus, gcd, inverse BigInt
	one := NewBigInt(1)
	for _, test := range modInverseTests {
		(&element).SetString(test.element, 10)
		(&modulus).SetString(test.modulus, 10)
		(&inverse).ModInverse(&element, &modulus)
		(&inverse).Mul(&inverse, &element)
		(&inverse).Mod(&inverse, &modulus)
		if (&inverse).Cmp(one) != 0 {
			t.Errorf("ModInverse(%d,%d)*%d%%%d=%d, not 1", &element, &modulus, &element, &modulus, &inverse)
		}
	}
	// exhaustive test for small values
	for n := 2; n < 100; n++ {
		(&modulus).SetInt64(int64(n))
		for x := 1; x < n; x++ {
			(&element).SetInt64(int64(x))
			(&gcd).GCD(nil, nil, &element, &modulus)
			if (&gcd).Cmp(one) != 0 {
				continue
			}
			(&inverse).ModInverse(&element, &modulus)
			(&inverse).Mul(&inverse, &element)
			(&inverse).Mod(&inverse, &modulus)
			if (&inverse).Cmp(one) != 0 {
				t.Errorf("ModInverse(%d,%d)*%d%%%d=%d, not 1", &element, &modulus, &element, &modulus, &inverse)
			}
		}
	}
}

// testModSqrt is a helper for TestModSqrt,
// which checks that ModSqrt can compute a square-root of elt^2.
func testModSqrt(t *testing.T, elt, mod, sq, sqrt *BigInt) bool {
	var sqChk, sqrtChk, sqrtsq BigInt
	sq.Mul(elt, elt)
	sq.Mod(sq, mod)
	z := sqrt.ModSqrt(sq, mod)
	if z != sqrt {
		t.Errorf("ModSqrt returned wrong value %s", z)
	}

	// test ModSqrt arguments outside the range [0,mod)
	sqChk.Add(sq, mod)
	z = sqrtChk.ModSqrt(&sqChk, mod)
	if z != &sqrtChk || z.Cmp(sqrt) != 0 {
		t.Errorf("ModSqrt returned inconsistent value %s", z)
	}
	sqChk.Sub(sq, mod)
	z = sqrtChk.ModSqrt(&sqChk, mod)
	if z != &sqrtChk || z.Cmp(sqrt) != 0 {
		t.Errorf("ModSqrt returned inconsistent value %s", z)
	}

	// test x aliasing z
	z = sqrtChk.ModSqrt(sqrtChk.Set(sq), mod)
	if z != &sqrtChk || z.Cmp(sqrt) != 0 {
		t.Errorf("ModSqrt returned inconsistent value %s", z)
	}

	// make sure we actually got a square root
	if sqrt.Cmp(elt) == 0 {
		return true // we found the "desired" square root
	}
	sqrtsq.Mul(sqrt, sqrt) // make sure we found the "other" one
	sqrtsq.Mod(&sqrtsq, mod)
	return sq.Cmp(&sqrtsq) == 0
}

var primes = []string{
	"2",
	"3",
	"5",
	"7",
	"11",

	"13756265695458089029",
	"13496181268022124907",
	"10953742525620032441",
	"17908251027575790097",

	// https://golang.org/issue/638
	"18699199384836356663",

	"98920366548084643601728869055592650835572950932266967461790948584315647051443",
	"94560208308847015747498523884063394671606671904944666360068158221458669711639",

	// https://primes.utm.edu/lists/small/small3.html
	"449417999055441493994709297093108513015373787049558499205492347871729927573118262811508386655998299074566974373711472560655026288668094291699357843464363003144674940345912431129144354948751003607115263071543163",
	"230975859993204150666423538988557839555560243929065415434980904258310530753006723857139742334640122533598517597674807096648905501653461687601339782814316124971547968912893214002992086353183070342498989426570593",
	"5521712099665906221540423207019333379125265462121169655563495403888449493493629943498064604536961775110765377745550377067893607246020694972959780839151452457728855382113555867743022746090187341871655890805971735385789993",
	"203956878356401977405765866929034577280193993314348263094772646453283062722701277632936616063144088173312372882677123879538709400158306567338328279154499698366071906766440037074217117805690872792848149112022286332144876183376326512083574821647933992961249917319836219304274280243803104015000563790123",

	// ECC primes: https://tools.ietf.org/html/draft-ladd-safecurves-02
	"3618502788666131106986593281521497120414687020801267626233049500247285301239",                                                                                  // Curve1174: 2^251-9
	"57896044618658097711785492504343953926634992332820282019728792003956564819949",                                                                                 // Curve25519: 2^255-19
	"9850501549098619803069760025035903451269934817616361666987073351061430442874302652853566563721228910201656997576599",                                           // E-382: 2^382-105
	"42307582002575910332922579714097346549017899709713998034217522897561970639123926132812109468141778230245837569601494931472367",                                 // Curve41417: 2^414-17
	"6864797660130609714981900799081393217269435300143305409394463459185543183397656052122559640661454554977296311391480858037121987999716643812574028291115057151", // E-521: 2^521-1
}

func TestBigIntModSqrt(t *testing.T) {
	var elt, mod, modx4, sq, sqrt BigInt
	r := rand.New(rand.NewSource(9))
	for i, s := range primes[1:] { // skip 2, use only odd primes
		mod.SetString(s, 10)
		modx4.Lsh(&mod, 2)

		// test a few random elements per prime
		for x := 1; x < 5; x++ {
			elt.Rand(r, &modx4)
			elt.Sub(&elt, &mod) // test range [-mod, 3*mod)
			if !testModSqrt(t, &elt, &mod, &sq, &sqrt) {
				t.Errorf("#%d: failed (sqrt(e) = %s)", i, &sqrt)
			}
		}

		if testing.Short() && i > 2 {
			break
		}
	}

	if testing.Short() {
		return
	}

	// exhaustive test for small values
	for n := 3; n < 100; n++ {
		mod.SetInt64(int64(n))
		if !mod.ProbablyPrime(10) {
			continue
		}
		isSquare := make([]bool, n)

		// test all the squares
		for x := 1; x < n; x++ {
			elt.SetInt64(int64(x))
			if !testModSqrt(t, &elt, &mod, &sq, &sqrt) {
				t.Errorf("#%d: failed (sqrt(%d,%d) = %s)", x, &elt, &mod, &sqrt)
			}
			isSquare[sq.Uint64()] = true
		}

		// test all non-squares
		for x := 1; x < n; x++ {
			sq.SetInt64(int64(x))
			z := sqrt.ModSqrt(&sq, &mod)
			if !isSquare[x] && z != nil {
				t.Errorf("#%d: failed (sqrt(%d,%d) = nil)", x, &sqrt, &mod)
			}
		}
	}
}

func TestBigIntIssue2607(t *testing.T) {
	// This code sequence used to hang.
	n := NewBigInt(10)
	n.Rand(rand.New(rand.NewSource(9)), n)
}

func TestBigIntSqrt(t *testing.T) {
	root := 0
	r := new(BigInt)
	for i := 0; i < 10000; i++ {
		if (root+1)*(root+1) <= i {
			root++
		}
		n := NewBigInt(int64(i))
		r.SetInt64(-2)
		r.Sqrt(n)
		if r.Cmp(NewBigInt(int64(root))) != 0 {
			t.Errorf("Sqrt(%v) = %v, want %v", n, r, root)
		}
	}

	for i := 0; i < 1000; i += 10 {
		n, _ := new(BigInt).SetString("1"+strings.Repeat("0", i), 10)
		r := new(BigInt).Sqrt(n)
		root, _ := new(BigInt).SetString("1"+strings.Repeat("0", i/2), 10)
		if r.Cmp(root) != 0 {
			t.Errorf("Sqrt(1e%d) = %v, want 1e%d", i, r, i/2)
		}
	}

	// Test aliasing.
	r.SetInt64(100)
	r.Sqrt(r)
	if r.Int64() != 10 {
		t.Errorf("Sqrt(100) = %v, want 10 (aliased output)", r.Int64())
	}
}

// We can't test this together with the other Exp tests above because
// it requires a different receiver setup.
func TestBigIntIssue22830(t *testing.T) {
	one := new(BigInt).SetInt64(1)
	base, _ := new(BigInt).SetString("84555555300000000000", 10)
	mod, _ := new(BigInt).SetString("66666670001111111111", 10)
	want, _ := new(BigInt).SetString("17888885298888888889", 10)

	var tests = []int64{
		0, 1, -1,
	}

	for _, n := range tests {
		m := NewBigInt(n)
		if got := m.Exp(base, one, mod); got.Cmp(want) != 0 {
			t.Errorf("(%v).Exp(%s, 1, %s) = %s, want %s", n, base, mod, got, want)
		}
	}
}

//
// Tests from src/math/big/intconv_test.go
//

var stringTests = []struct {
	in   string
	out  string
	base int
	val  int64
	ok   bool
}{
	// invalid inputs
	{in: ""},
	{in: "a"},
	{in: "z"},
	{in: "+"},
	{in: "-"},
	{in: "0b"},
	{in: "0o"},
	{in: "0x"},
	{in: "0y"},
	{in: "2", base: 2},
	{in: "0b2", base: 0},
	{in: "08"},
	{in: "8", base: 8},
	{in: "0xg", base: 0},
	{in: "g", base: 16},

	// invalid inputs with separators
	// (smoke tests only - a comprehensive set of tests is in natconv_test.go)
	{in: "_"},
	{in: "0_"},
	{in: "_0"},
	{in: "-1__0"},
	{in: "0x10_"},
	{in: "1_000", base: 10}, // separators are not permitted for bases != 0
	{in: "d_e_a_d", base: 16},

	// valid inputs
	{"0", "0", 0, 0, true},
	{"0", "0", 10, 0, true},
	{"0", "0", 16, 0, true},
	{"+0", "0", 0, 0, true},
	{"-0", "0", 0, 0, true},
	{"10", "10", 0, 10, true},
	{"10", "10", 10, 10, true},
	{"10", "10", 16, 16, true},
	{"-10", "-10", 16, -16, true},
	{"+10", "10", 16, 16, true},
	{"0b10", "2", 0, 2, true},
	{"0o10", "8", 0, 8, true},
	{"0x10", "16", 0, 16, true},
	{in: "0x10", base: 16},
	{"-0x10", "-16", 0, -16, true},
	{"+0x10", "16", 0, 16, true},
	{"00", "0", 0, 0, true},
	{"0", "0", 8, 0, true},
	{"07", "7", 0, 7, true},
	{"7", "7", 8, 7, true},
	{"023", "19", 0, 19, true},
	{"23", "23", 8, 19, true},
	{"cafebabe", "cafebabe", 16, 0xcafebabe, true},
	{"0b0", "0", 0, 0, true},
	{"-111", "-111", 2, -7, true},
	{"-0b111", "-7", 0, -7, true},
	{"0b1001010111", "599", 0, 0x257, true},
	{"1001010111", "1001010111", 2, 0x257, true},
	{"A", "a", 36, 10, true},
	{"A", "A", 37, 36, true},
	{"ABCXYZ", "abcxyz", 36, 623741435, true},
	{"ABCXYZ", "ABCXYZ", 62, 33536793425, true},

	// valid input with separators
	// (smoke tests only - a comprehensive set of tests is in natconv_test.go)
	{"1_000", "1000", 0, 1000, true},
	{"0b_1010", "10", 0, 10, true},
	{"+0o_660", "432", 0, 0660, true},
	{"-0xF00D_1E", "-15731998", 0, -0xf00d1e, true},
}

func TestBigIntText(t *testing.T) {
	z := new(BigInt)
	for _, test := range stringTests {
		if !test.ok {
			continue
		}

		_, ok := z.SetString(test.in, test.base)
		if !ok {
			t.Errorf("%v: failed to parse", test)
			continue
		}

		base := test.base
		if base == 0 {
			base = 10
		}

		if got := z.Text(base); got != test.out {
			t.Errorf("%v: got %s; want %s", test, got, test.out)
		}
	}
}

func TestBigIntAppendText(t *testing.T) {
	z := new(BigInt)
	var buf []byte
	for _, test := range stringTests {
		if !test.ok {
			continue
		}

		_, ok := z.SetString(test.in, test.base)
		if !ok {
			t.Errorf("%v: failed to parse", test)
			continue
		}

		base := test.base
		if base == 0 {
			base = 10
		}

		i := len(buf)
		buf = z.Append(buf, base)
		if got := string(buf[i:]); got != test.out {
			t.Errorf("%v: got %s; want %s", test, got, test.out)
		}
	}
}

func TestBigIntGetString(t *testing.T) {
	format := func(base int) string {
		switch base {
		case 2:
			return "%b"
		case 8:
			return "%o"
		case 16:
			return "%x"
		}
		return "%d"
	}

	z := new(BigInt)
	for i, test := range stringTests {
		if !test.ok {
			continue
		}
		z.SetInt64(test.val)

		if test.base == 10 {
			if got := z.String(); got != test.out {
				t.Errorf("#%da got %s; want %s", i, got, test.out)
			}
		}

		f := format(test.base)
		got := fmt.Sprintf(f, z)
		if f == "%d" {
			if got != fmt.Sprintf("%d", test.val) {
				t.Errorf("#%db got %s; want %d", i, got, test.val)
			}
		} else {
			if got != test.out {
				t.Errorf("#%dc got %s; want %s", i, got, test.out)
			}
		}
	}
}

func TestBigIntSetString(t *testing.T) {
	tmp := new(BigInt)
	for i, test := range stringTests {
		// initialize to a non-zero value so that issues with parsing
		// 0 are detected
		tmp.SetInt64(1234567890)
		n1, ok1 := new(BigInt).SetString(test.in, test.base)
		n2, ok2 := tmp.SetString(test.in, test.base)
		expected := NewBigInt(test.val)
		if ok1 != test.ok || ok2 != test.ok {
			t.Errorf("#%d (input '%s') ok incorrect (should be %t)", i, test.in, test.ok)
			continue
		}
		if !ok1 {
			if n1 != nil {
				t.Errorf("#%d (input '%s') n1 != nil", i, test.in)
			}
			continue
		}
		if !ok2 {
			if n2 != nil {
				t.Errorf("#%d (input '%s') n2 != nil", i, test.in)
			}
			continue
		}

		if n1.Cmp(expected) != 0 {
			t.Errorf("#%d (input '%s') got: %s want: %d", i, test.in, n1, test.val)
		}
		if n2.Cmp(expected) != 0 {
			t.Errorf("#%d (input '%s') got: %s want: %d", i, test.in, n2, test.val)
		}
	}
}

var formatTests = []struct {
	input  string
	format string
	output string
}{
	{"<nil>", "%x", "<nil>"},
	{"<nil>", "%#x", "<nil>"},
	{"<nil>", "%#y", "%!y(big.Int=<nil>)"},

	{"10", "%b", "1010"},
	{"10", "%o", "12"},
	{"10", "%d", "10"},
	{"10", "%v", "10"},
	{"10", "%x", "a"},
	{"10", "%X", "A"},
	{"-10", "%X", "-A"},
	{"10", "%y", "%!y(big.Int=10)"},
	{"-10", "%y", "%!y(big.Int=-10)"},

	{"10", "%#b", "0b1010"},
	{"10", "%#o", "012"},
	{"10", "%O", "0o12"},
	{"-10", "%#b", "-0b1010"},
	{"-10", "%#o", "-012"},
	{"-10", "%O", "-0o12"},
	{"10", "%#d", "10"},
	{"10", "%#v", "10"},
	{"10", "%#x", "0xa"},
	{"10", "%#X", "0XA"},
	{"-10", "%#X", "-0XA"},
	{"10", "%#y", "%!y(big.Int=10)"},
	{"-10", "%#y", "%!y(big.Int=-10)"},

	{"1234", "%d", "1234"},
	{"1234", "%3d", "1234"},
	{"1234", "%4d", "1234"},
	{"-1234", "%d", "-1234"},
	{"1234", "% 5d", " 1234"},
	{"1234", "%+5d", "+1234"},
	{"1234", "%-5d", "1234 "},
	{"1234", "%x", "4d2"},
	{"1234", "%X", "4D2"},
	{"-1234", "%3x", "-4d2"},
	{"-1234", "%4x", "-4d2"},
	{"-1234", "%5x", " -4d2"},
	{"-1234", "%-5x", "-4d2 "},
	{"1234", "%03d", "1234"},
	{"1234", "%04d", "1234"},
	{"1234", "%05d", "01234"},
	{"1234", "%06d", "001234"},
	{"-1234", "%06d", "-01234"},
	{"1234", "%+06d", "+01234"},
	{"1234", "% 06d", " 01234"},
	{"1234", "%-6d", "1234  "},
	{"1234", "%-06d", "1234  "},
	{"-1234", "%-06d", "-1234 "},

	{"1234", "%.3d", "1234"},
	{"1234", "%.4d", "1234"},
	{"1234", "%.5d", "01234"},
	{"1234", "%.6d", "001234"},
	{"-1234", "%.3d", "-1234"},
	{"-1234", "%.4d", "-1234"},
	{"-1234", "%.5d", "-01234"},
	{"-1234", "%.6d", "-001234"},

	{"1234", "%8.3d", "    1234"},
	{"1234", "%8.4d", "    1234"},
	{"1234", "%8.5d", "   01234"},
	{"1234", "%8.6d", "  001234"},
	{"-1234", "%8.3d", "   -1234"},
	{"-1234", "%8.4d", "   -1234"},
	{"-1234", "%8.5d", "  -01234"},
	{"-1234", "%8.6d", " -001234"},

	{"1234", "%+8.3d", "   +1234"},
	{"1234", "%+8.4d", "   +1234"},
	{"1234", "%+8.5d", "  +01234"},
	{"1234", "%+8.6d", " +001234"},
	{"-1234", "%+8.3d", "   -1234"},
	{"-1234", "%+8.4d", "   -1234"},
	{"-1234", "%+8.5d", "  -01234"},
	{"-1234", "%+8.6d", " -001234"},

	{"1234", "% 8.3d", "    1234"},
	{"1234", "% 8.4d", "    1234"},
	{"1234", "% 8.5d", "   01234"},
	{"1234", "% 8.6d", "  001234"},
	{"-1234", "% 8.3d", "   -1234"},
	{"-1234", "% 8.4d", "   -1234"},
	{"-1234", "% 8.5d", "  -01234"},
	{"-1234", "% 8.6d", " -001234"},

	{"1234", "%.3x", "4d2"},
	{"1234", "%.4x", "04d2"},
	{"1234", "%.5x", "004d2"},
	{"1234", "%.6x", "0004d2"},
	{"-1234", "%.3x", "-4d2"},
	{"-1234", "%.4x", "-04d2"},
	{"-1234", "%.5x", "-004d2"},
	{"-1234", "%.6x", "-0004d2"},

	{"1234", "%8.3x", "     4d2"},
	{"1234", "%8.4x", "    04d2"},
	{"1234", "%8.5x", "   004d2"},
	{"1234", "%8.6x", "  0004d2"},
	{"-1234", "%8.3x", "    -4d2"},
	{"-1234", "%8.4x", "   -04d2"},
	{"-1234", "%8.5x", "  -004d2"},
	{"-1234", "%8.6x", " -0004d2"},

	{"1234", "%+8.3x", "    +4d2"},
	{"1234", "%+8.4x", "   +04d2"},
	{"1234", "%+8.5x", "  +004d2"},
	{"1234", "%+8.6x", " +0004d2"},
	{"-1234", "%+8.3x", "    -4d2"},
	{"-1234", "%+8.4x", "   -04d2"},
	{"-1234", "%+8.5x", "  -004d2"},
	{"-1234", "%+8.6x", " -0004d2"},

	{"1234", "% 8.3x", "     4d2"},
	{"1234", "% 8.4x", "    04d2"},
	{"1234", "% 8.5x", "   004d2"},
	{"1234", "% 8.6x", "  0004d2"},
	{"1234", "% 8.7x", " 00004d2"},
	{"1234", "% 8.8x", " 000004d2"},
	{"-1234", "% 8.3x", "    -4d2"},
	{"-1234", "% 8.4x", "   -04d2"},
	{"-1234", "% 8.5x", "  -004d2"},
	{"-1234", "% 8.6x", " -0004d2"},
	{"-1234", "% 8.7x", "-00004d2"},
	{"-1234", "% 8.8x", "-000004d2"},

	{"1234", "%-8.3d", "1234    "},
	{"1234", "%-8.4d", "1234    "},
	{"1234", "%-8.5d", "01234   "},
	{"1234", "%-8.6d", "001234  "},
	{"1234", "%-8.7d", "0001234 "},
	{"1234", "%-8.8d", "00001234"},
	{"-1234", "%-8.3d", "-1234   "},
	{"-1234", "%-8.4d", "-1234   "},
	{"-1234", "%-8.5d", "-01234  "},
	{"-1234", "%-8.6d", "-001234 "},
	{"-1234", "%-8.7d", "-0001234"},
	{"-1234", "%-8.8d", "-00001234"},

	{"16777215", "%b", "111111111111111111111111"}, // 2**24 - 1

	{"0", "%.d", ""},
	{"0", "%.0d", ""},
	{"0", "%3.d", ""},
}

func TestBigIntFormat(t *testing.T) {
	for i, test := range formatTests {
		var x *BigInt
		if test.input != "<nil>" {
			var ok bool
			x, ok = new(BigInt).SetString(test.input, 0)
			if !ok {
				t.Errorf("#%d failed reading input %s", i, test.input)
			}
		}
		output := fmt.Sprintf(test.format, x)
		if output != test.output {
			t.Errorf("#%d got %q; want %q, {%q, %q, %q}", i, output, test.output, test.input, test.format, test.output)
		}
	}
}

var scanTests = []struct {
	input     string
	format    string
	output    string
	remaining int
}{
	{"1010", "%b", "10", 0},
	{"0b1010", "%v", "10", 0},
	{"12", "%o", "10", 0},
	{"012", "%v", "10", 0},
	{"10", "%d", "10", 0},
	{"10", "%v", "10", 0},
	{"a", "%x", "10", 0},
	{"0xa", "%v", "10", 0},
	{"A", "%X", "10", 0},
	{"-A", "%X", "-10", 0},
	{"+0b1011001", "%v", "89", 0},
	{"0xA", "%v", "10", 0},
	{"0 ", "%v", "0", 1},
	{"2+3", "%v", "2", 2},
	{"0XABC 12", "%v", "2748", 3},
}

func TestBigIntScan(t *testing.T) {
	var buf bytes.Buffer
	for i, test := range scanTests {
		x := new(BigInt)
		buf.Reset()
		buf.WriteString(test.input)
		if _, err := fmt.Fscanf(&buf, test.format, x); err != nil {
			t.Errorf("#%d error: %s", i, err)
		}
		if x.String() != test.output {
			t.Errorf("#%d got %s; want %s", i, x.String(), test.output)
		}
		if buf.Len() != test.remaining {
			t.Errorf("#%d got %d bytes remaining; want %d", i, buf.Len(), test.remaining)
		}
	}
}

//
// Tests from src/math/big/intmarsh_test.go
//

var encodingTests = []string{
	"0",
	"1",
	"2",
	"10",
	"1000",
	"1234567890",
	"298472983472983471903246121093472394872319615612417471234712061",
}

func TestBigIntGobEncoding(t *testing.T) {
	var medium bytes.Buffer
	enc := gob.NewEncoder(&medium)
	dec := gob.NewDecoder(&medium)
	for _, test := range encodingTests {
		for _, sign := range []string{"", "+", "-"} {
			x := sign + test
			medium.Reset() // empty buffer for each test case (in case of failures)
			var tx BigInt
			tx.SetString(x, 10)
			if err := enc.Encode(&tx); err != nil {
				t.Errorf("encoding of %s failed: %s", &tx, err)
				continue
			}
			var rx BigInt
			if err := dec.Decode(&rx); err != nil {
				t.Errorf("decoding of %s failed: %s", &tx, err)
				continue
			}
			if rx.Cmp(&tx) != 0 {
				t.Errorf("transmission of %s failed: got %s want %s", &tx, &rx, &tx)
			}
		}
	}
}

// Sending a nil BigInt pointer (inside a slice) on a round trip through gob should yield a zero.
func TestBigIntGobEncodingNilIntInSlice(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	dec := gob.NewDecoder(buf)

	var in = make([]*BigInt, 1)
	err := enc.Encode(&in)
	if err != nil {
		t.Errorf("gob encode failed: %q", err)
	}
	var out []*BigInt
	err = dec.Decode(&out)
	if err != nil {
		t.Fatalf("gob decode failed: %q", err)
	}
	if len(out) != 1 {
		t.Fatalf("wrong len; want 1 got %d", len(out))
	}
	var zero BigInt
	if out[0].Cmp(&zero) != 0 {
		t.Fatalf("transmission of (*BigInt)(nil) failed: got %s want 0", out)
	}
}

func TestBigIntJSONEncoding(t *testing.T) {
	for _, test := range encodingTests {
		for _, sign := range []string{"", "+", "-"} {
			x := sign + test
			var tx BigInt
			tx.SetString(x, 10)
			b, err := json.Marshal(&tx)
			if err != nil {
				t.Errorf("marshaling of %s failed: %s", &tx, err)
				continue
			}
			var rx BigInt
			if err := json.Unmarshal(b, &rx); err != nil {
				t.Errorf("unmarshaling of %s failed: %s", &tx, err)
				continue
			}
			if rx.Cmp(&tx) != 0 {
				t.Errorf("JSON encoding of %s failed: got %s want %s", &tx, &rx, &tx)
			}
		}
	}
}

func TestBigIntXMLEncoding(t *testing.T) {
	for _, test := range encodingTests {
		for _, sign := range []string{"", "+", "-"} {
			x := sign + test
			var tx BigInt
			tx.SetString(x, 0)
			b, err := xml.Marshal(&tx)
			if err != nil {
				t.Errorf("marshaling of %s failed: %s", &tx, err)
				continue
			}
			var rx BigInt
			if err := xml.Unmarshal(b, &rx); err != nil {
				t.Errorf("unmarshaling of %s failed: %s", &tx, err)
				continue
			}
			if rx.Cmp(&tx) != 0 {
				t.Errorf("XML encoding of %s failed: got %s want %s", &tx, &rx, &tx)
			}
		}
	}
}

//
// Benchmarks from src/math/big/int_test.go
//

func BenchmarkBigIntBinomial(b *testing.B) {
	var z BigInt
	for i := b.N - 1; i >= 0; i-- {
		z.Binomial(1000, 990)
	}
}

func BenchmarkBigIntQuoRem(b *testing.B) {
	x, _ := new(BigInt).SetString("153980389784927331788354528594524332344709972855165340650588877572729725338415474372475094155672066328274535240275856844648695200875763869073572078279316458648124537905600131008790701752441155668003033945258023841165089852359980273279085783159654751552359397986180318708491098942831252291841441726305535546071", 0)
	y, _ := new(BigInt).SetString("7746362281539803897849273317883545285945243323447099728551653406505888775727297253384154743724750941556720663282745352402758568446486952008757638690735720782793164586481245379056001310087907017524411556680030339452580238411650898523599802732790857831596547515523593979861803187084910989428312522918414417263055355460715745539358014631136245887418412633787074173796862711588221766398229333338511838891484974940633857861775630560092874987828057333663969469797013996401149696897591265769095952887917296740109742927689053276850469671231961384715398038978492733178835452859452433234470997285516534065058887757272972533841547437247509415567206632827453524027585684464869520087576386907357207827931645864812453790560013100879070175244115566800303394525802384116508985235998027327908578315965475155235939798618031870849109894283125229184144172630553554607112725169432413343763989564437170644270643461665184965150423819594083121075825", 0)
	q := new(BigInt)
	r := new(BigInt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.QuoRem(y, x, r)
	}
}

func BenchmarkBigIntExp(b *testing.B) {
	x, _ := new(BigInt).SetString("11001289118363089646017359372117963499250546375269047542777928006103246876688756735760905680604646624353196869572752623285140408755420374049317646428185270079555372763503115646054602867593662923894140940837479507194934267532831694565516466765025434902348314525627418515646588160955862839022051353653052947073136084780742729727874803457643848197499548297570026926927502505634297079527299004267769780768565695459945235586892627059178884998772989397505061206395455591503771677500931269477503508150175717121828518985901959919560700853226255420793148986854391552859459511723547532575574664944815966793196961286234040892865", 0)
	y, _ := new(BigInt).SetString("0xAC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF72", 0)
	n, _ := new(BigInt).SetString("0xAC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF73", 0)
	out := new(BigInt)
	for i := 0; i < b.N; i++ {
		out.Exp(x, y, n)
	}
}

func BenchmarkBigIntExp2(b *testing.B) {
	x, _ := new(BigInt).SetString("2", 0)
	y, _ := new(BigInt).SetString("0xAC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF72", 0)
	n, _ := new(BigInt).SetString("0xAC6BDB41324A9A9BF166DE5E1389582FAF72B6651987EE07FC3192943DB56050A37329CBB4A099ED8193E0757767A13DD52312AB4B03310DCD7F48A9DA04FD50E8083969EDB767B0CF6095179A163AB3661A05FBD5FAAAE82918A9962F0B93B855F97993EC975EEAA80D740ADBF4FF747359D041D5C33EA71D281E446B14773BCA97B43A23FB801676BD207A436C6481F1D2B9078717461A5B9D32E688F87748544523B524B0D57D5EA77A2775D2ECFA032CFBDBF52FB3786160279004E57AE6AF874E7303CE53299CCC041C7BC308D82A5698F3A8D0C38271AE35F8E9DBFBB694B5C803D89F7AE435DE236D525F54759B65E372FCD68EF20FA7111F9E4AFF73", 0)
	out := new(BigInt)
	for i := 0; i < b.N; i++ {
		out.Exp(x, y, n)
	}
}

func BenchmarkBigIntBitset(b *testing.B) {
	z := new(BigInt)
	z.SetBit(z, 512, 1)
	b.ResetTimer()
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		z.SetBit(z, i&512, 1)
	}
}

func BenchmarkBigIntBitsetNeg(b *testing.B) {
	z := NewBigInt(-1)
	z.SetBit(z, 512, 0)
	b.ResetTimer()
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		z.SetBit(z, i&512, 0)
	}
}

func BenchmarkBigIntBitsetOrig(b *testing.B) {
	z := new(BigInt)
	altSetBit(z, z, 512, 1)
	b.ResetTimer()
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		altSetBit(z, z, i&512, 1)
	}
}

func BenchmarkBigIntBitsetNegOrig(b *testing.B) {
	z := NewBigInt(-1)
	altSetBit(z, z, 512, 0)
	b.ResetTimer()
	b.StartTimer()
	for i := b.N - 1; i >= 0; i-- {
		altSetBit(z, z, i&512, 0)
	}
}

func BenchmarkBigIntModInverse(b *testing.B) {
	p := new(BigInt).SetInt64(1) // Mersenne prime 2**1279 -1
	p.Lsh(p, 1279)
	p.Sub(p, bigOne)
	x := new(BigInt).Sub(p, bigOne)
	z := new(BigInt)
	for i := 0; i < b.N; i++ {
		z.ModInverse(x, p)
	}
}

func BenchmarkBigIntSqrt(b *testing.B) {
	n, _ := new(BigInt).SetString("1"+strings.Repeat("0", 1001), 10)
	b.ResetTimer()
	t := new(BigInt)
	for i := 0; i < b.N; i++ {
		t.Sqrt(n)
	}
}

// randBigInt returns a pseudo-random Int in the range [1<<(size-1), (1<<size) - 1].
func randBigInt(r *rand.Rand, size uint) *BigInt {
	n := new(BigInt).Lsh(bigOne, size-1)
	x := new(BigInt).Rand(r, n)
	return x.Add(x, n) // make sure result > 1<<(size-1)
}

func benchmarkBigIntDiv(b *testing.B, aSize, bSize int) {
	var r = rand.New(rand.NewSource(1234))
	aa := randBigInt(r, uint(aSize))
	bb := randBigInt(r, uint(bSize))
	if aa.Cmp(bb) < 0 {
		aa, bb = bb, aa
	}
	x := new(BigInt)
	y := new(BigInt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.DivMod(aa, bb, y)
	}
}

func BenchmarkBigIntDiv(b *testing.B) {
	sizes := []int{
		10, 20, 50, 100, 200, 500, 1000,
		1e4, 1e5, 1e6, 1e7,
	}
	for _, i := range sizes {
		j := 2 * i
		b.Run(fmt.Sprintf("%d/%d", j, i), func(b *testing.B) {
			benchmarkBigIntDiv(b, j, i)
		})
	}
}

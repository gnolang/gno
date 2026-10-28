// Copyright 2026 The Gno Authors. All rights reserved.
// Use of this source code is governed by the Gno LICENSE at the root of
// the repository.

package goconst

import (
	"go/token"
	"math/big"
)

// A Meter is told each unit of work an operation is about to do, before
// it runs: the representation dispatched on and the size of its
// operands. It knows nothing about gas; the caller prices the work. A
// nil Meter charges nothing.
type Meter interface {
	Charge(w Work)
}

// Work is one operation on one representation.
type Work struct {
	Op    token.Token // the operator, or token.ILLEGAL for a conversion
	Conv  Conv        // the conversion, when Op is token.ILLEGAL
	Rep   Rep         // the representation the operands are matched to
	Bits  int         // the wider operand: bit length (a rational: numerator plus denominator), or bytes of a string
	Shift uint        // the count of a shift
	Unary bool        // a unary operation (Op is token.SUB, ADD, XOR or NOT)
}

// Rep is the representation an operation dispatches on, in matching
// order: an operation on two representations runs on the greater.
type Rep uint8

const (
	RepNone   Rep = iota
	RepBool       // boolVal
	RepInt64      // int64Val: a machine word
	RepInt        // intVal: *big.Int
	RepRat        // ratVal: *big.Rat, components below 4096 bits
	RepFloat      // floatVal: *big.Float, 512 bits
	RepString     // *stringVal
)

// Conv is a conversion metered on its own.
type Conv uint8

const (
	ConvNone    Conv = iota
	ConvToInt        // ToInt: a Float to an Int
	ConvToFloat      // ToFloat: an Int to a Float
	ConvFloat32      // Float32Val
	ConvFloat64      // Float64Val
	ConvPromote      // Make of a *big.Rat with a component of 4096 bits or more: a 512-bit division
	ConvString       // StringVal: a concatenation flattened
)

func repOf(x Value) (Rep, int) {
	switch x := x.(type) {
	case boolVal:
		return RepBool, 0
	case int64Val:
		return RepInt64, 0
	case intVal:
		return RepInt, x.val.BitLen()
	case ratVal:
		return RepRat, ratBits(x.val)
	case floatVal:
		return RepFloat, 0
	case *stringVal:
		return RepString, len(x.string())
	}
	return RepNone, 0
}

func ratBits(r *big.Rat) int { return r.Num().BitLen() + r.Denom().BitLen() }

// chargeBinary charges m for op on x and y as they will be matched: the
// greater representation, the wider operand.
func chargeBinary(m Meter, op token.Token, x, y Value) {
	if m == nil {
		return
	}
	xr, xb := repOf(x)
	yr, yb := repOf(y)
	m.Charge(Work{Op: op, Rep: max(xr, yr), Bits: max(xb, yb)})
}

func chargeUnary(m Meter, op token.Token, x Value) {
	if m == nil {
		return
	}
	r, b := repOf(x)
	m.Charge(Work{Op: op, Rep: r, Bits: b, Unary: true})
}

func chargeConv(m Meter, c Conv, x Value) {
	if m == nil {
		return
	}
	r, b := repOf(x)
	m.Charge(Work{Conv: c, Rep: r, Bits: b})
}

package gnolang

import (
	"fmt"
	"math"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/softfloat"
)

const (
	mask  = 0x7FF
	shift = 64 - 11 - 1
	bias  = 1023
)

type (
	SoftFloat64 uint64
	SoftFloat32 uint32
)

func Trunc(x SoftFloat64) SoftFloat64 {
	cmp, _ := softfloat.Fcmp64(uint64(x), softfloat.Fintto64(0))
	if _, _, _, isInf, IsNaN := softfloat.Funpack64(uint64(x)); cmp == 0 || isInf || IsNaN {
		return x
	}

	d, _ := Modf(x)
	return d
}

func Modf(u SoftFloat64) (it SoftFloat64, frac SoftFloat64) {
	if u.Le(1) {
		switch {
		case u.Lt(0):
			it, frac = Modf(u.Neg())
			return -it, -frac
		case u.Eq(0):
			return u, u // Return -0, -0 when f == -0
		}
		return 0, u
	}

	it = u
	e := uint(it>>shift)&mask - bias

	// Keep the top 12+e bits, the integer part; clear the rest.
	if e < 64-12 {
		it &^= 1<<(64-12-e) - 1
	}

	frac = u.Sub(it)
	return
}

func ConvertToSoftFloat64(n any) SoftFloat64 {
	switch n := n.(type) {
	case SoftFloat64:
		return n
	case SoftFloat32:
		return SoftFloat64(softfloat.F32to64(uint32(n)))
	case int:
		return SoftFloat64(softfloat.Fintto64(int64(n)))
	case int32:
		return SoftFloat64(softfloat.Fint32to64(n))
	case int8:
		return SoftFloat64(softfloat.Fint32to64(int32(n)))
	case int16:
		return SoftFloat64(softfloat.Fint32to64(int32(n)))
	case int64:
		return SoftFloat64(softfloat.Fint64to64(n))
	case uint:
		return SoftFloat64(softfloat.Fuint64to64(uint64(n)))
	case uint16:
		return SoftFloat64(softfloat.Fuint64to64(uint64(n)))
	case uint32:
		return SoftFloat64(softfloat.Fuint64to64(uint64(n)))
	case uint8:
		return SoftFloat64(softfloat.Fuint64to64(uint64(n)))
	case uint64:
		return SoftFloat64(softfloat.Fuint64to64(n))
	case float32:
		return SoftFloat32(math.Float32bits(n)).SoftFloat64()
	case float64:
		return SoftFloat64(math.Float64bits(n))
	default:
		panic(fmt.Sprintf("unsupported type: %T", n))
	}
}

func ConvertToSoftFloat32(n any) SoftFloat32 {
	switch n := n.(type) {
	case SoftFloat64:
		return SoftFloat32(softfloat.F64to32(uint64(n)))
	case SoftFloat32:
		return n
	case int:
		return SoftFloat32(softfloat.Fintto32(int64(n)))
	case int32:
		return SoftFloat32(softfloat.Fint32to32(n))
	case int8:
		return SoftFloat32(softfloat.Fint32to32(int32(n)))
	case int16:
		return SoftFloat32(softfloat.Fint32to32(int32(n)))
	case int64:
		return SoftFloat32(softfloat.Fint64to32(n))
	case uint:
		return SoftFloat32(softfloat.Fuint64to32(uint64(n)))
	case uint16:
		return SoftFloat32(softfloat.Fuint64to32(uint64(n)))
	case uint32:
		return SoftFloat32(softfloat.Fuint64to32(uint64(n)))
	case uint8:
		return SoftFloat32(softfloat.Fuint64to32(uint64(n)))
	case uint64:
		return SoftFloat32(softfloat.Fuint64to32(n))
	case float32:
		return SoftFloat32(math.Float32bits(n))
	case float64:
		return SoftFloat64(math.Float64bits(n)).SoftFloat32()
	default:
		panic(fmt.Sprintf("unsupported type: %T", n))
	}
}

// SoftFloat64

func (f SoftFloat64) String() string {
	return fmt.Sprintf("%v", math.Float64frombits(uint64(f)))
}

func (f SoftFloat64) Float64() float64 {
	return math.Float64frombits(uint64(f))
}

func (f SoftFloat64) Float32() float32 {
	return float32(math.Float64frombits(uint64(f)))
}

func (f SoftFloat64) SoftFloat32() SoftFloat32 {
	return SoftFloat32(softfloat.F64to32(uint64(f)))
}

func (f SoftFloat64) Int() int {
	n, _ := softfloat.F64toint(uint64(f))
	return int(n)
}

func (f SoftFloat64) Int64() int64 {
	return softfloat.F64toint64(uint64(f))
}

func (f SoftFloat64) Int32() int32 {
	return softfloat.F64toint32(uint64(f))
}

func (f SoftFloat64) Int16() int16 {
	return int16(softfloat.F64toint32(uint64(f)))
}

func (f SoftFloat64) Int8() int8 {
	return int8(softfloat.F64toint32(uint64(f)))
}

func (f SoftFloat64) Uint() uint {
	return uint(softfloat.F64touint64(uint64(f)))
}

func (f SoftFloat64) Uint64() uint64 {
	return softfloat.F64touint64(uint64(f))
}

func (f SoftFloat64) Uint32() uint32 {
	return uint32(softfloat.F64touint64(uint64(f)))
}

func (f SoftFloat64) Uint16() uint16 {
	return uint16(softfloat.F64touint64(uint64(f)))
}

func (f SoftFloat64) Uint8() uint8 {
	return uint8(softfloat.F64touint64(uint64(f)))
}

func (f SoftFloat64) Add(g any) SoftFloat64 {
	return SoftFloat64(softfloat.Fadd64(uint64(f), uint64(ConvertToSoftFloat64(g))))
}

func (f SoftFloat64) Sub(g any) SoftFloat64 {
	return SoftFloat64(softfloat.Fsub64(uint64(f), uint64(ConvertToSoftFloat64(g))))
}

func (f SoftFloat64) Mul(g any) SoftFloat64 {
	return SoftFloat64(softfloat.Fmul64(uint64(f), uint64(ConvertToSoftFloat64(g))))
}

func (f SoftFloat64) Div(g any) SoftFloat64 {
	return SoftFloat64(softfloat.Fdiv64(uint64(f), uint64(ConvertToSoftFloat64(g))))
}

func (f SoftFloat64) Neg() SoftFloat64 {
	return SoftFloat64(softfloat.Fneg64(uint64(f)))
}

func (f SoftFloat64) Trunc() SoftFloat64 {
	return Trunc(f)
}

// ==
func (f SoftFloat64) Eq(g any) bool {
	return softfloat.Feq64(uint64(f), uint64(ConvertToSoftFloat64(g)))
}

// >
func (f SoftFloat64) Gt(g any) bool {
	return softfloat.Fgt64(uint64(f), uint64(ConvertToSoftFloat64(g)))
}

// >=
func (f SoftFloat64) Ge(g any) bool {
	return softfloat.Fge64(uint64(f), uint64(ConvertToSoftFloat64(g)))
}

// <
func (f SoftFloat64) Lt(g any) bool {
	return !softfloat.Fge64(uint64(f), uint64(ConvertToSoftFloat64(g)))
}

// <=
func (f SoftFloat64) Le(g any) bool {
	return !softfloat.Fgt64(uint64(f), uint64(ConvertToSoftFloat64(g)))
}

// SoftFloat32

func (f SoftFloat32) Float32() float32 {
	return math.Float32frombits(uint32(f))
}

func (f SoftFloat32) Float64() float64 {
	return math.Float64frombits(softfloat.F32to64(uint32(f)))
}

func (f SoftFloat32) SoftFloat64() SoftFloat64 {
	return SoftFloat64(softfloat.F32to64(uint32(f)))
}

func (f SoftFloat32) Int() int {
	return int(softfloat.F32toint64(uint32(f)))
}

func (f SoftFloat32) Int64() int64 {
	return softfloat.F32toint64(uint32(f))
}

func (f SoftFloat32) Int32() int32 {
	return softfloat.F32toint32(uint32(f))
}

func (f SoftFloat32) Int16() int16 {
	return int16(softfloat.F32toint32(uint32(f)))
}

func (f SoftFloat32) Int8() int8 {
	return int8(softfloat.F32toint32(uint32(f)))
}

func (f SoftFloat32) Uint() uint {
	return uint(softfloat.F32touint64(uint32(f)))
}

func (f SoftFloat32) Uint64() uint64 {
	return softfloat.F32touint64(uint32(f))
}

func (f SoftFloat32) Uint32() uint32 {
	return uint32(softfloat.F32touint64(uint32(f)))
}

func (f SoftFloat32) Uint16() uint16 {
	return uint16(softfloat.F32touint64(uint32(f)))
}

func (f SoftFloat32) Uint8() uint8 {
	return uint8(softfloat.F32touint64(uint32(f)))
}

func (f SoftFloat32) Add(g any) SoftFloat32 {
	return SoftFloat32(softfloat.Fadd32(uint32(f), uint32(ConvertToSoftFloat32(g))))
}

func (f SoftFloat32) Sub(g any) SoftFloat32 {
	return SoftFloat32(softfloat.Fsub32(uint32(f), uint32(ConvertToSoftFloat32(g))))
}

func (f SoftFloat32) Mul(g any) SoftFloat32 {
	return SoftFloat32(softfloat.Fmul32(uint32(f), uint32(ConvertToSoftFloat32(g))))
}

func (f SoftFloat32) Div(g any) SoftFloat32 {
	return SoftFloat32(softfloat.Fdiv32(uint32(f), uint32(ConvertToSoftFloat32(g))))
}

func (f SoftFloat32) Neg() SoftFloat32 {
	return SoftFloat32(softfloat.Fneg32(uint32(f)))
}

func (f SoftFloat32) Trunc() SoftFloat32 {
	return SoftFloat32(softfloat.F64to32(uint64(Trunc(SoftFloat64(softfloat.F32to64(uint32(f)))))))
}

// ==
func (f SoftFloat32) Eq(g any) bool {
	return softfloat.Feq32(uint32(f), uint32(ConvertToSoftFloat32(g)))
}

// >
func (f SoftFloat32) Gt(g any) bool {
	return softfloat.Fgt32(uint32(f), uint32(ConvertToSoftFloat32(g)))
}

// >=
func (f SoftFloat32) Ge(g any) bool {
	return softfloat.Fge32(uint32(f), uint32(ConvertToSoftFloat32(g)))
}

// <
func (f SoftFloat32) Lt(g any) bool {
	return !softfloat.Fge32(uint32(f), uint32(ConvertToSoftFloat32(g)))
}

// <=
func (f SoftFloat32) Le(g any) bool {
	return !softfloat.Fgt32(uint32(f), uint32(ConvertToSoftFloat32(g)))
}

func (f SoftFloat32) String() string {
	return fmt.Sprintf("%v", math.Float32frombits(uint32(f)))
}

// Package softfloat is a copy of the Go runtime's softfloat64.go file.
// It is a pure software floating point implementation. It can be used to
// perform determinstic, hardware-independent floating point computations.
//
// This package uses shortnames to refer to its different operations. Here is a
// quick reference:
//
//	add   f + g
//	sub   f - g
//	mul   f * g
//	div   f / g
//	neg   (- f)
//	eq    f == g
//	gt    f > g
//	ge    f >= g
package softfloat

const (
	mask  = 0x7FF
	shift = 64 - 11 - 1
	bias  = 1023
)

type Float64 uint64
type Float32 uint32

func Trunc(x Float64) Float64 {
	_, _, _, isInf, IsNaN := funpack64(uint64(x))
	if x == 0 || isInf || IsNaN {
		return x
	}

	d, _ := Modf(x)
	return d
}
func Modf(u Float64) (int Float64, frac Float64) {
	cmp, _ := Fcmp64(u, 1)

	if cmp < 0 {
		cmp, _ := Fcmp64(u, 0)
		switch {
		case cmp < 0:
			int, frac = Modf(Fneg64(u))
			return Fneg64(int), Fneg64(frac)
		case cmp == 0:
			return u, u // Return -0, -0 when f == -0
		}
		return 0, u
	}

	e := uint(u>>shift)&mask - bias

	// Keep the top 12+e bits, the integer part; clear the rest.
	if e < 64-12 {
		u &^= 1<<(64-12-e) - 1
	}

	frac = Fsub64(u, int)
	return
}

// This file mostly exports the functions from runtime_softfloat64.go

//go:generate sh copy.sh

func Fadd64(f, g Float64) Float64 { return Float64(fadd64(uint64(f), uint64(g))) }
func Fsub64(f, g Float64) Float64 { return Float64(fsub64(uint64(f), uint64(g))) }
func Fmul64(f, g Float64) Float64 { return Float64(fmul64(uint64(f), uint64(g))) }
func Fdiv64(f, g Float64) Float64 { return Float64(fdiv64(uint64(f), uint64(g))) }
func Fneg64(f Float64) Float64    { return Float64(fneg64(uint64(f))) }
func Feq64(f, g Float64) bool     { return feq64(uint64(f), uint64(g)) }
func Fgt64(f, g Float64) bool     { return fgt64(uint64(f), uint64(g)) }
func Fge64(f, g Float64) bool     { return fge64(uint64(f), uint64(g)) }

func Fadd32(f, g Float32) Float32 { return Float32(fadd32(uint32(f), uint32(g))) }
func Fsub32(f, g Float32) Float32 { return Float32(fadd32(uint32(f), uint32(Fneg32(uint32(g))))) }
func Fmul32(f, g Float32) Float32 { return Float32(fmul32(uint32(f), uint32(g))) }
func Fdiv32(f, g Float32) Float32 { return Float32(fdiv32(uint32(f), uint32(g))) }
func Feq32(f, g Float32) bool     { return feq32(uint32(f), uint32(g)) }
func Fgt32(f, g Float32) bool     { return fgt32(uint32(f), uint32(g)) }
func Fge32(f, g Float32) bool     { return fge32(uint32(f), uint32(g)) }

func Fcmp64(f, g Float64) (cmp int32, isnan bool) { return fcmp64(uint64(f), uint64(g)) }

func Fneg32(f uint32) uint32 {
	// Not defined in runtime - this is a copy similar to fneg64.
	return f ^ (1 << (mantbits32 + expbits32))
}

// Conversions

func Fintto64(val int64) Float64 { return Float64(fintto64(val)) }
func Fintto32(val int64) Float32 { return Float32(fintto32(val)) }

func F32to64(f Float32) Float64               { return Float64(f32to64(uint32(f))) }
func F32toint32(x Float32) int32              { return f32toint32(uint32(x)) }
func F32toint64(x Float32) int64              { return f32toint64(uint32(x)) }
func F32touint64(x Float32) uint64            { return f32touint64(uint32(x)) }
func F64to32(f Float64) Float32               { return Float32(f64to32(uint64(f))) }
func F64toint(f Float64) (val int64, ok bool) { return f64toint(uint64(f)) }
func F64toint32(x Float64) int32              { return f64toint32(uint64(x)) }
func F64toint64(x Float64) int64              { return f64toint64(uint64(x)) }
func F64touint64(x Float64) uint64            { return f64touint64(uint64(x)) }
func Fint32to32(x int32) Float32              { return Float32(fint32to32(x)) }
func Fint32to64(x int32) Float64              { return Float64(fint32to64(x)) }
func Fint64to32(x int64) Float32              { return Float32(fint64to32(x)) }
func Fint64to64(x int64) Float64              { return Float64(fint64to64(x)) }
func Fuint64to32(x uint64) Float32            { return Float32(fuint64to32(x)) }
func Fuint64to64(x uint64) Float64            { return Float64(fuint64to64(x)) }

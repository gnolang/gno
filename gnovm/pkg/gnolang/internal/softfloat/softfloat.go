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

// This file mostly exports the functions from runtime_softfloat64.go

//go:generate sh copy.sh

func Fadd64(f, g uint64) uint64 { return fadd64(f, g) }
func Fsub64(f, g uint64) uint64 { return fsub64(f, g) }
func Fmul64(f, g uint64) uint64 { return fmul64(f, g) }
func Fdiv64(f, g uint64) uint64 { return fdiv64(f, g) }
func Fneg64(f uint64) uint64    { return fneg64(f) }
func Feq64(f, g uint64) bool    { return feq64(f, g) }
func Fgt64(f, g uint64) bool    { return fgt64(f, g) }
func Fge64(f, g uint64) bool    { return fge64(f, g) }

func Fadd32(f, g uint32) uint32 { return fadd32(f, g) }
func Fsub32(f, g uint32) uint32 { return fadd32(f, Fneg32(g)) }
func Fmul32(f, g uint32) uint32 { return fmul32(f, g) }
func Fdiv32(f, g uint32) uint32 { return fdiv32(f, g) }
func Feq32(f, g uint32) bool    { return feq32(f, g) }
func Fgt32(f, g uint32) bool    { return fgt32(f, g) }
func Fge32(f, g uint32) bool    { return fge32(f, g) }

func Fcmp64(f, g uint64) (cmp int32, isnan bool) { return fcmp64(f, g) }

func Fneg32(f uint32) uint32 {
	// Not defined in runtime - this is a copy similar to fneg64.
	return f ^ (1 << (mantbits32 + expbits32))
}

// Conversions

func Fintto64(val int64) (f uint64) { return fintto64(val) }
func Fintto32(val int64) (f uint32) { return fintto32(val) }

func F32to64(f uint32) uint64                { return f32to64(f) }
func F32toint32(x uint32) int32              { return f32toint32(x) }
func F32toint64(x uint32) int64              { return f32toint64(x) }
func F32touint64(x uint32) uint64            { return f32touint64(x) }
func F64to32(f uint64) uint32                { return f64to32(f) }
func F64toint(f uint64) (val int64, ok bool) { return f64toint(f) }
func F64toint32(x uint64) int32              { return f64toint32(x) }
func F64toint64(x uint64) int64              { return f64toint64(x) }
func F64touint64(x uint64) uint64            { return f64touint64(x) }
func Fint32to32(x int32) uint32              { return fint32to32(x) }
func Fint32to64(x int32) uint64              { return fint32to64(x) }
func Fint64to32(x int64) uint32              { return fint64to32(x) }
func Fint64to64(x int64) uint64              { return fint64to64(x) }

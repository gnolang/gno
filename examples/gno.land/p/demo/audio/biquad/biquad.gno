package biquad

import "math"

type Filter struct {
	a1, a2, b0, b1, b2     float64
	x1_f, x2_f, y1_f, y2_f float64
}

func New(fc float64, fs float64, q float64, db float64, filter_type string) (f *Filter) {
	w0 := 2 * 3.1415926535897932384626 * (fc / fs)
	cosW := math.Cos(w0)
	sinW := math.Sin(w0)
	A := 1.0 // db = 0
	_ = A
	// A := math.Pow(10, (db / 40))
	alpha := sinW / (2 * q)
	beta := 1 // db = 0
	// beta := A ^ 0.5/q
	_ = beta
	b0 := (1 - cosW) / 2
	b1 := 1 - cosW
	b2 := (1 - cosW) / 2
	a0 := 1 + alpha
	a1 := -2 * cosW
	a2 := 1 - alpha
	if filter_type == "highpass" {
		b0 = (1 + cosW) / 2
		b1 = -(1 + cosW)
		b2 = (1 + cosW) / 2
	} else if filter_type == "lowpass" {
		// do nothing, default
	}

	b0 = b0 / a0
	b1 = b1 / a0
	b2 = b2 / a0
	a1 = a1 / a0
	a2 = a2 / a0

	return &Filter{
		a1: a1,
		a2: a2,
		b0: b0,
		b1: b1,
		b2: b2,
	}
}

func (f *Filter) Update(x float64) (y float64) {
	y = f.b0*x + f.b1*f.x1_f + f.b2*f.x2_f - f.a1*f.y1_f - f.a2*f.y2_f
	f.x2_f = f.x1_f
	f.x1_f = x
	f.y2_f = f.y1_f
	f.y1_f = y
	return
}

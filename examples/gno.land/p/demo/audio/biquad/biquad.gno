// biquad implements a standard digital biquad filter which is a
// set of equations that can be used to filter signals. More information
// at https://en.wikipedia.org/wiki/Digital_biquad_filter
// Currently it is implemented with only a highpass and a lowpass
// filter but it is straightforward to add others if nessecary.
package biquad

import "math"

type Filter struct {
	a1, a2, b0, b1, b2     float64
	x1_f, x2_f, y1_f, y2_f float64
}

// New provides a new biquad according to the equation for a first-order
// biquad. The algorithm follows the equations in this: https://www.ti.com/lit/pdf/slaa447
// Inputs:
//
//	fc: filter cutoff
//	fs: sample rate
//	q: resonance coefficient
//	db: shelf decibels (strength)
//	filterType: string that specifies the type of filter (currently only 'highpass' or 'lowpass')
//
// Outputs:
//
//	a filter object
func New(fc float64, fs float64, q float64, db float64, filterType string) *Filter {
	w0 := 2 * math.Pi * (fc / fs)
	cosW := math.Cos(w0)
	sinW := math.Sin(w0)
	alpha := sinW / (2 * q)
	b0 := (1 - cosW) / 2
	b1 := 1 - cosW
	b2 := (1 - cosW) / 2
	a0 := 1 + alpha
	a1 := -2 * cosW
	a2 := 1 - alpha
	if filterType == "highpass" {
		b0 = (1 + cosW) / 2
		b1 = -(1 + cosW)
		b2 = (1 + cosW) / 2
	} else if filterType == "lowpass" {
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

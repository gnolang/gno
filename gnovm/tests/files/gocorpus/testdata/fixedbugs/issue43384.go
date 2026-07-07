// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

type T int

func (T) Mv()  {}
func (*T) Mp() {}

type P1 struct{ T }
type P2 struct{ *T }
type P3 *struct{ T }
type P4 *struct{ *T }

func _() {
	{
		var p P1
		p.Mv()
		(&p).Mv()
		(*&p).Mv()
		p.Mp()
		(&p).Mp()
		(*&p).Mp()
	}
	{
		var p P2
		p.Mv()
		(&p).Mv()
		(*&p).Mv()
		p.Mp()
		(&p).Mp()
		(*&p).Mp()
	}
	{
		var p P3
		p.Mv()     // ERROR "undefined"
		(&p).Mv()  // ERROR "undefined"
		(*&p).Mv() // ERROR "undefined"
		(**&p).Mv()
		(*p).Mv()
		(&*p).Mv()
		p.Mp()     // ERROR "undefined"
		(&p).Mp()  // ERROR "undefined"
		(*&p).Mp() // ERROR "undefined"
		(**&p).Mp()
		(*p).Mp()
		(&*p).Mp()
	}
	{
		var p P4
		p.Mv()     // ERROR "undefined"
		(&p).Mv()  // ERROR "undefined"
		(*&p).Mv() // ERROR "undefined"
		(**&p).Mv()
		(*p).Mv()
		(&*p).Mv()
		p.Mp()     // ERROR "undefined"
		(&p).Mp()  // ERROR "undefined"
		(*&p).Mp() // ERROR "undefined"
		(**&p).Mp()
		(*p).Mp()
		(&*p).Mp()
	}
}

func _() {
	type P5 struct{ T }
	type P6 struct{ *T }
	type P7 *struct{ T }
	type P8 *struct{ *T }

	{
		var p P5
		p.Mv()
		(&p).Mv()
		(*&p).Mv()
		p.Mp()
		(&p).Mp()
		(*&p).Mp()
	}
	{
		var p P6
		p.Mv()
		(&p).Mv()
		(*&p).Mv()
		p.Mp()
		(&p).Mp()
		(*&p).Mp()
	}
	{
		var p P7
		p.Mv()     // ERROR "undefined"
		(&p).Mv()  // ERROR "undefined"
		(*&p).Mv() // ERROR "undefined"
		(**&p).Mv()
		(*p).Mv()
		(&*p).Mv()
		p.Mp()     // ERROR "undefined"
		(&p).Mp()  // ERROR "undefined"
		(*&p).Mp() // ERROR "undefined"
		(**&p).Mp()
		(*p).Mp()
		(&*p).Mp()
	}
	{
		var p P8
		p.Mv()     // ERROR "undefined"
		(&p).Mv()  // ERROR "undefined"
		(*&p).Mv() // ERROR "undefined"
		(**&p).Mv()
		(*p).Mv()
		(&*p).Mv()
		p.Mp()     // ERROR "undefined"
		(&p).Mp()  // ERROR "undefined"
		(*&p).Mp() // ERROR "undefined"
		(**&p).Mp()
		(*p).Mp()
		(&*p).Mp()
	}
}

// GoTypeCheckError:
// line 40: p.Mv undefined (type P3 has no field or method Mv)
// line 41: (&p).Mv undefined (type *P3 has no field or method Mv)
// line 42: (*&p).Mv undefined (type P3 has no field or method Mv)
// line 46: p.Mp undefined (type P3 has no field or method Mp)
// line 47: (&p).Mp undefined (type *P3 has no field or method Mp)
// line 48: (*&p).Mp undefined (type P3 has no field or method Mp)
// line 55: p.Mv undefined (type P4 has no field or method Mv)
// line 56: (&p).Mv undefined (type *P4 has no field or method Mv)
// line 57: (*&p).Mv undefined (type P4 has no field or method Mv)
// line 61: p.Mp undefined (type P4 has no field or method Mp)
// line 62: (&p).Mp undefined (type *P4 has no field or method Mp)
// line 63: (*&p).Mp undefined (type P4 has no field or method Mp)
// line 96: p.Mv undefined (type P7 has no field or method Mv)
// line 97: (&p).Mv undefined (type *P7 has no field or method Mv)
// line 98: (*&p).Mv undefined (type P7 has no field or method Mv)
// line 102: p.Mp undefined (type P7 has no field or method Mp)
// line 103: (&p).Mp undefined (type *P7 has no field or method Mp)
// line 104: (*&p).Mp undefined (type P7 has no field or method Mp)
// line 111: p.Mv undefined (type P8 has no field or method Mv)
// line 112: (&p).Mv undefined (type *P8 has no field or method Mv)
// line 113: (*&p).Mv undefined (type P8 has no field or method Mv)
// line 117: p.Mp undefined (type P8 has no field or method Mp)
// line 118: (&p).Mp undefined (type *P8 has no field or method Mp)
// line 119: (*&p).Mp undefined (type P8 has no field or method Mp)

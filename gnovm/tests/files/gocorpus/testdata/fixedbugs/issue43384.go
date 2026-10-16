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

// GnoError:
// line 40: missing field Mv in gno.land/p/filetest/p.P3
// line 41: should not happen
// line 42: missing field Mv in gno.land/p/filetest/p.P3
// line 46: missing field Mp in gno.land/p/filetest/p.P3
// line 47: should not happen
// line 48: missing field Mp in gno.land/p/filetest/p.P3
// line 55: missing field Mv in gno.land/p/filetest/p.P4
// line 56: should not happen
// line 57: missing field Mv in gno.land/p/filetest/p.P4
// line 61: missing field Mp in gno.land/p/filetest/p.P4
// line 62: should not happen
// line 63: missing field Mp in gno.land/p/filetest/p.P4
// line 96: missing field Mv in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P7
// line 97: should not happen
// line 98: missing field Mv in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P7
// line 102: missing field Mp in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P7
// line 103: should not happen
// line 104: missing field Mp in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P7
// line 111: missing field Mv in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P8
// line 112: should not happen
// line 113: missing field Mv in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P8
// line 117: missing field Mp in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P8
// line 118: should not happen
// line 119: missing field Mp in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P8

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 41, 47, 56, 62, 97, 103, 112, 118; internal errors: lines 41, 47, 56, 62, 97, 103, 112, 118

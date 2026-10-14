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

// GoTypeCheckError:
// line 40
// line 41
// line 42
// line 46
// line 47
// line 48
// line 55
// line 56
// line 57
// line 61
// line 62
// line 63
// line 96
// line 97
// line 98
// line 102
// line 103
// line 104
// line 111
// line 112
// line 113
// line 117
// line 118
// line 119

// GnoWordingDiffers:
// line 40: gc expects: undefined; gno: missing field Mv in gno.land/p/filetest/p.P3
// line 41: gc expects: undefined; gno: should not happen
// line 42: gc expects: undefined; gno: missing field Mv in gno.land/p/filetest/p.P3
// line 46: gc expects: undefined; gno: missing field Mp in gno.land/p/filetest/p.P3
// line 47: gc expects: undefined; gno: should not happen
// line 48: gc expects: undefined; gno: missing field Mp in gno.land/p/filetest/p.P3
// line 55: gc expects: undefined; gno: missing field Mv in gno.land/p/filetest/p.P4
// line 56: gc expects: undefined; gno: should not happen
// line 57: gc expects: undefined; gno: missing field Mv in gno.land/p/filetest/p.P4
// line 61: gc expects: undefined; gno: missing field Mp in gno.land/p/filetest/p.P4
// line 62: gc expects: undefined; gno: should not happen
// line 63: gc expects: undefined; gno: missing field Mp in gno.land/p/filetest/p.P4
// line 96: gc expects: undefined; gno: missing field Mv in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P7
// line 97: gc expects: undefined; gno: should not happen
// line 98: gc expects: undefined; gno: missing field Mv in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P7
// line 102: gc expects: undefined; gno: missing field Mp in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P7
// line 103: gc expects: undefined; gno: should not happen
// line 104: gc expects: undefined; gno: missing field Mp in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P7
// line 111: gc expects: undefined; gno: missing field Mv in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P8
// line 112: gc expects: undefined; gno: should not happen
// line 113: gc expects: undefined; gno: missing field Mv in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P8
// line 117: gc expects: undefined; gno: missing field Mp in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P8
// line 118: gc expects: undefined; gno: should not happen
// line 119: gc expects: undefined; gno: missing field Mp in gno.land/p/filetest/p[gno.land/p/filetest/p/issue43384.go:71:1-125:2].P8

// Verdict: Clean errorcheck

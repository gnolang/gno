// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import "time"

type T int

func (T) Mv()  {}
func (*T) Mp() {}

var _ = []int{
	T.Mv,    // ERROR "cannot use T\.Mv|incompatible type"
	(*T).Mv, // ERROR "cannot use \(\*T\)\.Mv|incompatible type"
	(*T).Mp, // ERROR "cannot use \(\*T\)\.Mp|incompatible type"

	time.Time.GobEncode,    // ERROR "cannot use time\.Time\.GobEncode|incompatible type"
	(*time.Time).GobEncode, // ERROR "cannot use \(\*time\.Time\)\.GobEncode|incompatible type"
	(*time.Time).GobDecode, // ERROR "cannot use \(\*time\.Time\)\.GobDecode|incompatible type"

}

// GnoError:
// line 17: cannot use func(gno.land/p/filetest/p.T) as int
// line 18: cannot use func(gno.land/p/filetest/p.T) as int
// line 19: cannot use func(*gno.land/p/filetest/p.T) as int
// line 21: cannot use func(time.Time) ([]uint8, .uverse.error) as int
// line 22: cannot use func(time.Time) ([]uint8, .uverse.error) as int
// line 23: cannot use func(*time.Time, []uint8) .uverse.error as int

// Verdict: Clean errorcheck

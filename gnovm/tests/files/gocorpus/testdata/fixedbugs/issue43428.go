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
// line 16: 2: cannot use func(gno.land/p/filetest/p.T) as int

// GoTypeCheckError:
// line 17
// line 18
// line 19
// line 21
// line 22
// line 23

// GnoWordingDiffers:
// line 17: gc expects: cannot use T\.Mv|incompatible type; gno: 2: cannot use func(gno.land/p/filetest/p.T) as int
// line 18: gc expects: cannot use \(\*T\)\.Mv|incompatible type; gno: 2: cannot use func(gno.land/p/filetest/p.T) as int
// line 19: gc expects: cannot use \(\*T\)\.Mp|incompatible type; gno: 2: cannot use func(gno.land/p/filetest/p.T) as int
// line 21: gc expects: cannot use time\.Time\.GobEncode|incompatible type; gno: 2: cannot use func(gno.land/p/filetest/p.T) as int
// line 22: gc expects: cannot use \(\*time\.Time\)\.GobEncode|incompatible type; gno: 2: cannot use func(gno.land/p/filetest/p.T) as int
// line 23: gc expects: cannot use \(\*time\.Time\)\.GobDecode|incompatible type; gno: 2: cannot use func(gno.land/p/filetest/p.T) as int

// Verdict: Clean errorcheck

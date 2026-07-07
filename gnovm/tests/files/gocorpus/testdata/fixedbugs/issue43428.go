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
// line 17: expected declaration, found T
// line 18: expected declaration, found '('
// line 19: expected declaration, found '('
// line 21: expected declaration, found time
// line 22: expected declaration, found '('
// line 23: expected declaration, found '('
// line 25: expected declaration, found '}'

// GoTypeCheckError:
// line 17: cannot use T.Mv (value of type func(T)) as int value in array or slice literal
// line 18: cannot use (*T).Mv (value of type func(*T)) as int value in array or slice literal
// line 19: cannot use (*T).Mp (value of type func(*T)) as int value in array or slice literal
// line 21: cannot use time.Time.GobEncode (value of type func(time.Time) ([]byte, error)) as int value in array or slice literal
// line 22: cannot use (*time.Time).GobEncode (value of type func(*time.Time) ([]byte, error)) as int value in array or slice literal
// line 23: cannot use (*time.Time).GobDecode (value of type func(t *time.Time, data []byte) error) as int value in array or slice literal

// GnoOverStrictError:
// line 16: 2: cannot use func(gno.land/p/filetest/p.T) as int
// line 25: expected declaration, found '}'

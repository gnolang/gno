// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type it struct {
	Floats bool
	inner  string
}

func main() {
	i1 := it{Floats: true}
	if i1.floats { // ERROR "(type it .* field or method floats, but does have field Floats)|undefined field or method"
	}
	i2 := &it{floats: false} // ERROR "cannot refer to unexported field floats in struct literal|unknown field|declared and not used"
	_ = &it{InneR: "foo"}    // ERROR "(but does have field inner)|unknown field"
	_ = i2
}

// GnoError:
// line 16: missing field floats in main.it
// line 18: expected declaration, found i2
// line 19: expected declaration, found _

// GoTypeCheckError:
// line 16: i1.floats undefined (type it has no field or method floats, but does have field Floats)
// line 18: unknown field floats in struct literal of type it, but does have Floats
// line 19: unknown field InneR in struct literal of type it, but does have inner

// errorcheck -goexperiment fieldtrack

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f(interface{ m() }) {}
func g()                 { f(new(T)) } // ERROR "m method is marked 'nointerface'"

type T struct{}

//go:nointerface
func (*T) m() {}

// Unsupported: Gno doesn't implement gc's //go:nointerface directive check; it accepts this file (gc rejects it), so there's no error to pin.

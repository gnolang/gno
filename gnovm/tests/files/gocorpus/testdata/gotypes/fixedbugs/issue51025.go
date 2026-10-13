// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ interface{ m() } = struct /* ERROR "m is a field, not a method" */ {
	m func()
}{}

var _ interface{ m() } = & /* ERROR "m is a field, not a method" */ struct {
	m func()
}{}

var _ interface{ M() } = struct /* ERROR "missing method M" */ {
	m func()
}{}

var _ interface{ M() } = & /* ERROR "missing method M" */ struct {
	m func()
}{}

// test case from issue
type I interface{ m() }
type T struct{ m func() }
type M struct{}

func (M) m() {}

func _() {
	var t T
	var m M
	var i I

	i = m
	i = t // ERROR "m is a field, not a method"
	_ = i
}

// GnoError:
// line 7: 4: struct{m func()} does not implement interface {m func()} (struct{m func()}.m is a field, not a method)
// line 11: 4: *struct{m func()} does not implement interface {m func()} (*struct{m func()}.m is a field, not a method)
// line 15: 4: struct{m func()} does not implement interface {M func()} (missing method M)
// line 19: 4: *struct{m func()} does not implement interface {M func()} (missing method M)
// line 36: gno.land/p/filetest/p.T does not implement gno.land/p/filetest/p.I (gno.land/p/filetest/p.T.m is a field, not a method)

// GoTypeCheckError:
// line 7
// line 11
// line 15
// line 19
// line 36

// Clean: errorcheck

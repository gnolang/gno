// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Used to crash
// https://golang.org/issue/204

package main

func () x()	// ERROR "no receiver"

func (a b, c d) x()	// ERROR "multiple receiver"

type b int

// GnoError:
// line 12: method has no receiver

// ParseBlocked:
// line 14: gc expects: multiple receiver

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 14

// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// When computing method sets with shadowed methods, make sure we
// compute whether a method promotion involved a pointer traversal
// based on the promoted method, not the shadowed method.

package main

import (
	"bytes"
	"fmt"
)

type mystruct struct {
	f int
}

func (t mystruct) String() string {
	return "FAIL"
}

func main() {
	type deep struct {
		mystruct
	}
	s := struct {
		deep
		*bytes.Buffer
	}{
		deep{},
		bytes.NewBufferString("ok"),
	}

	if got := s.String(); got != "ok" {
		panic(got)
	}

	var i fmt.Stringer = s
	if got := i.String(); got != "ok" {
		panic(got)
	}
}


// Fixing: PR #5721 (fix/method40, BFS lookup); verified clean on branch, broken on master; re-golden after merge.

// GnoOutput:

// GnoError:
// main/issue24547.go:38:12-20: missing field String in struct{deep main[main/issue24547.go:26:1-46:2].deep; Buffer *bytes.Buffer}

// GoOutput:

// KnownIssue:
// Embedded method lookup mishandles shadowing across depths: Buffer's
// promoted String (depth 1) should shadow mystruct's (depth 2), but Gno
// rejects the selector outright ("missing field String"). Same root cause
// as fixedbugs/bug485.go.

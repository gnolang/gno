// compile -dynlink

//go:build 386 || amd64 || arm || arm64 || ppc64le || s390x
// (platforms that support -dynlink flag)

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 58826: assembler cannot handle global access with large
// offset in -dynlink mode on ARM64.

package p

var x [2197]uint8

func F() {
	for _, i := range x {
		G(i)
	}
}

func G(uint8)

// GnoPreprocessError:
// line 23: function G does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoBuildError:
// line 23

// Verdict: Clean compile

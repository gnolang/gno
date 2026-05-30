// compile

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "runtime"

func f(x interface{}) {
	runtime.KeepAlive(x)
}

// GoTypeCheckError:
// line 12: undefined: runtime.KeepAlive

// KnownIssue:
// line 12: name KeepAlive not declared

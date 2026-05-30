// errorcheck -d=panic

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "bytes"

type _ struct{ bytes.nonexist } // ERROR "unexported|undefined"

type _ interface{ bytes.nonexist } // ERROR "unexported|undefined|expected signature or type name"

func main() {
	var _ bytes.Buffer
	var _ bytes.buffer // ERROR "unexported|undefined"
}

// GnoError:
// line 11: cannot access bytes.nonexist from main
// line 13: cannot access bytes.nonexist from main
// line 17: cannot access bytes.buffer from main

// GoTypeCheckError:
// line 11: undefined: bytes.nonexist
// line 13: undefined: bytes.nonexist
// line 17: undefined: bytes.buffer (but have Buffer)

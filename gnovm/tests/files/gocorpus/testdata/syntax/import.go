// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io",	// ERROR "unexpected comma"
	"os"
)

// GnoError:
// line 10: expected ';', found ','
// line 11: "os" imported and not used

// GnoWordingDiffers:
// line 10: gc expects: unexpected comma; gno: expected ';', found ','

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 11

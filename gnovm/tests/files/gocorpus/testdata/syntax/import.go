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

// GoTypeCheckError:
// line 10: expected ';', found ','

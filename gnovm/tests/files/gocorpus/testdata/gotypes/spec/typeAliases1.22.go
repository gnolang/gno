// -lang=go1.22

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package aliasTypes

type _ = int
type _[P /* ERROR "generic type alias requires go1.23 or later" */ any] = int

// GoTypeCheckError:
// line 10

// GoTypesOnly:
// line 10: gc expects: generic type alias requires go1.23 or later

// Verdict: Clean errorcheck

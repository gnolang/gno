// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package A

type (
	_ = A
	A /* ERROR "invalid recursive type: A refers to itself" */ = A
)

// GnoError:
// line 8: name not defined: A
// line 9: invalid recursive type: A -> A

// GoTypeCheckError:
// line 9

// GnoOverStrictError:
// line 8: name not defined: A

// KnownIssue:
// errorcheck: 1 over-strict line(s), 0 uncaught marker(s)

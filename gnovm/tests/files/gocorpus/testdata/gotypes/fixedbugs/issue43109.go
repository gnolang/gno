// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Ensure there is no "imported and not used" error
// if a package wasn't imported in the first place.

package p

import . "/foo" // ERROR "could not import /foo"

// GnoError:
// line 10: import path cannot be absolute path: "/foo"

// GoTypeCheckError:
// line 10

// GnoWordingDiffers:
// line 10: gc expects: could not import /foo; gno: import path cannot be absolute path: "/foo"

// Verdict: Clean errorcheck

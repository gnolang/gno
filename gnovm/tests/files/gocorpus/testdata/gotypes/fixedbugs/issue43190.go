// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The errors below are produced by the parser, but we check
// them here for consistency with the types2 tests.

package p

import ; // ERROR "missing import path"
import "" // ERROR "invalid import path (empty string)"
import
var /* ERROR "missing import path" */ _ int
import .; // ERROR "missing import path"
import 'x' // ERROR "import path must be a string"
var _ int
import /* ERROR "imports must appear before other declarations" */ _ "math"

// Don't repeat previous error for each immediately following import ...
import ()
import (.) // ERROR "missing import path"
import (
	"fmt"
	.
) // ERROR "missing import path"

// ... but remind with error again if we start a new import section after
// other declarations
var _ = fmt.Println
import /* ERROR "imports must appear before other declarations" */ _ "math"
import _ "math"

// GnoError:
// line 10: missing import path

// GoTypeCheckError:
// line 10

// ParseBlocked:
// line 11: gc expects: invalid import path (empty string)
// line 13: gc expects: missing import path
// line 14: gc expects: missing import path
// line 15: gc expects: import path must be a string
// line 17: gc expects: imports must appear before other declarations
// line 21: gc expects: missing import path
// line 25: gc expects: missing import path
// line 30: gc expects: imports must appear before other declarations

// Verdict: Clean errorcheck

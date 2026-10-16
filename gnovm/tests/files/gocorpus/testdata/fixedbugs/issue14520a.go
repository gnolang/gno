// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f

import /* // ERROR "import path" */ `
bogus`

// GnoError:
// line 9: invalid import path "\nbogus": invalid character '\n'

// Verdict: KnownIssue errorcheck: #6344; on-chain path accepts the file

// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f

import /* // ERROR "import path" */ `
bogus`

// GnoError:
// line 9: 7: unknown import path 
// bogus

// GoTypeCheckError:
// line 9: invalid import path (invalid character U+000A)

// errorcheck

// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

import "fmt"  // GC_ERROR "imported and not used"

const n = fmt // ERROR "fmt without selector|unexpected reference to package|use of package fmt not in selector"

// GnoError:
// line 11: package fmt cannot only be referred to in a selector expression

// GoTypeCheckError:
// line 9: "fmt" imported and not used
// line 11: use of package fmt not in selector

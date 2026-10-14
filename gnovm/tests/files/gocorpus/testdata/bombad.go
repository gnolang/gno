// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Here for reference, but hard to test automatically
// because the BOM muddles the
// processing done by ../run.

package main

func main() {
	﻿// There's a bom here.	// ERROR "BOM"
	//﻿ And here.	// ERROR "BOM"
	/*﻿ And here.*/	// ERROR "BOM"
	println("hi﻿ there") // and here	// ERROR "BOM"
}

// GnoError:
// line 14: illegal byte order mark
// line 15: illegal byte order mark
// line 16: illegal byte order mark
// line 17: illegal byte order mark
// line 18: expected '}', found 'EOF'

// GoTypeCheckError:
// line 14

// GnoOverStrictError:
// line 15: illegal byte order mark
// line 18: expected '}', found 'EOF'

// GnoWordingDiffers:
// line 14: gc expects: BOM; gno: illegal byte order mark
// line 16: gc expects: BOM; gno: illegal byte order mark
// line 17: gc expects: BOM; gno: illegal byte order mark

// Verdict: KnownIssue errorcheck: 2 over-strict line(s), 0 uncaught marker(s)

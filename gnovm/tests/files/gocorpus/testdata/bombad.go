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
// line 14: illegal byte order mark (and 4 more errors)

// GoTypeCheckError:
// line 14

// UncaughtError:
// line 16: uncaught; gc expects: BOM
// line 17: uncaught; gc expects: BOM

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 2 uncaught marker(s)

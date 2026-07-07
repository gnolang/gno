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
// line 15: illegal byte order mark (and 2 more errors)
// line 16: illegal byte order mark (and 1 more errors)
// line 17: illegal byte order mark

// GoTypeCheckError:
// line 14: illegal byte order mark (and 4 more errors)
// line 16: illegal byte order mark (and 1 more errors)
// line 17: illegal byte order mark

// GnoOverStrictError:
// line 15: illegal byte order mark (and 2 more errors)

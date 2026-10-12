// errorcheck -d=panic

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

func main() {	// ERROR "package"
}

// UncaughtError:
// line 7: uncaught; gc expects: package

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 1 uncaught marker(s)

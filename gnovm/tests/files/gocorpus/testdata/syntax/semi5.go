// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main()
{	// ERROR "unexpected semicolon or newline before .?{.?"

// GnoError:
// line 9: function main does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 10: unexpected semicolon or newline before {

// GoTypeCheckError:
// line 10: unexpected semicolon or newline before {

// GnoOverStrictError:
// line 9: function main does not have a body but is not natively defined (did you build after pulling from the repository?)

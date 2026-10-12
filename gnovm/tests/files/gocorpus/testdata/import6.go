// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that invalid imports are rejected by the compiler.
// Does not compile.

package main

// Each of these pairs tests both `` vs "" strings
// and also use of invalid characters spelled out as
// escape sequences and written directly.
// For example `"\x00"` tests import "\x00"
// while "`\x00`" tests import `<actual-NUL-byte>`.
import ""         // ERROR "import path"
import ``         // ERROR "import path"
import "\x00"     // ERROR "import path"
import `\x00`     // ERROR "import path"
import "\x7f"     // ERROR "import path"
import `\x7f`     // ERROR "import path"
import "a!"       // ERROR "import path"
import `a!`       // ERROR "import path"
import "a b"      // ERROR "import path"
import `a b`      // ERROR "import path"
import "a\\b"     // ERROR "import path"
import `a\\b`     // ERROR "import path"
import "\"`a`\""  // ERROR "import path"
import `\"a\"`    // ERROR "import path"
import "\x80\x80" // ERROR "import path"
import `\x80\x80` // ERROR "import path"
import "\xFFFD"   // ERROR "import path"
import `\xFFFD`   // ERROR "import path"

// Invalid local imports.
// types2 adds extra "not used" error.
import "/foo"  // ERROR "import path cannot be absolute path|not used"
import "c:/foo"  // ERROR "import path contains invalid character|invalid character"

// GnoError:
// line 17: invalid zero package path in testStore().pkgGetter:
// --- preprocess stack ---
// stack 1: file main/import6.go:10:1-39:16
// ------------------------

// GoTypeCheckError:
// line 17
// line 18
// line 19
// line 20
// line 21
// line 22
// line 23
// line 24
// line 25
// line 26
// line 27
// line 28
// line 29
// line 30
// line 31
// line 32
// line 33
// line 34
// line 38
// line 39

// GoTypesOnly:
// line 18: gc expects: import path
// line 19: gc expects: import path
// line 20: gc expects: import path
// line 21: gc expects: import path
// line 22: gc expects: import path
// line 23: gc expects: import path
// line 24: gc expects: import path
// line 25: gc expects: import path
// line 26: gc expects: import path
// line 27: gc expects: import path
// line 28: gc expects: import path
// line 29: gc expects: import path
// line 30: gc expects: import path
// line 31: gc expects: import path
// line 32: gc expects: import path
// line 33: gc expects: import path
// line 34: gc expects: import path
// line 38: gc expects: import path cannot be absolute path|not used
// line 39: gc expects: import path contains invalid character|invalid character

// Verdict: Clean errorcheck

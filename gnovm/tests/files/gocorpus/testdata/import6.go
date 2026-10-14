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
// line 17: invalid import path (empty string)
// line 18: invalid import path (empty string)
// line 19: invalid import path "\x00": invalid character '\x00'
// line 20: invalid import path "\\x00": invalid character '\\'
// line 21: invalid import path "\x7f": invalid character '\x7f'
// line 22: invalid import path "\\x7f": invalid character '\\'
// line 23: invalid import path "a!": invalid character '!'
// line 24: invalid import path "a!": invalid character '!'
// line 25: invalid import path "a b": invalid character ' '
// line 26: invalid import path "a b": invalid character ' '
// line 27: invalid import path "a\\b": invalid character '\\'
// line 28: invalid import path "a\\\\b": invalid character '\\'
// line 29: invalid import path "\"`a`\"": invalid character '"'
// line 30: invalid import path "\\\"a\\\"": invalid character '\\'
// line 31: import path contains invalid UTF-8 character: "\x80\x80"
// line 32: invalid import path "\\x80\\x80": invalid character '\\'
// line 33: import path contains invalid UTF-8 character: "\xffFD"
// line 34: invalid import path "\\xFFFD": invalid character '\\'
// line 38: import path cannot be absolute path: "/foo"
// line 39: invalid import path "c:/foo": invalid character ':'

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

// Verdict: Clean errorcheck

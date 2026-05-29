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
// line 19: invalid import path (invalid character U+0000)
// line 20: unknown import path \x00
// line 21: unknown import path 
// line 22: unknown import path \x7f
// line 23: unknown import path a!
// line 24: unknown import path a!
// line 25: unknown import path a b
// line 26: unknown import path a b
// line 27: unknown import path a\b
// line 28: unknown import path a\\b
// line 29: unknown import path "`a`"
// line 30: unknown import path \"a\"
// line 31: unknown import path €€
// line 32: unknown import path \x80\x80
// line 33: unknown import path ÿFD
// line 34: unknown import path \xFFFD
// line 38: unknown import path /foo
// line 39: invalid import path (invalid character U+003A ':')

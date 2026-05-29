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
// line 17: invalid zero package path in testStore().pkgGetter
// line 18: invalid zero package path in testStore().pkgGetter

// GoTypeCheckError:
// line 19: invalid import path (invalid character U+0000)
// line 20: invalid import path (invalid character U+005C '\')
// line 21: invalid import path (invalid character U+007F)
// line 22: invalid import path (invalid character U+005C '\')
// line 23: invalid import path (invalid character U+0021 '!')
// line 24: invalid import path (invalid character U+0021 '!')
// line 25: invalid import path (invalid character U+0020 ' ')
// line 26: invalid import path (invalid character U+0020 ' ')
// line 27: invalid import path (invalid character U+005C '\')
// line 28: invalid import path (invalid character U+005C '\')
// line 29: invalid import path (invalid character U+0022 '"')
// line 30: invalid import path (invalid character U+005C '\')
// line 31: invalid import path (invalid character U+FFFD '�')
// line 32: invalid import path (invalid character U+005C '\')
// line 33: invalid import path (invalid character U+FFFD '�')
// line 34: invalid import path (invalid character U+005C '\')
// line 38: could not import /foo (unknown import path "/foo")
// line 39: invalid import path (invalid character U+003A ':')

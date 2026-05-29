// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func init() // ERROR "missing function body|cannot declare init"

// GnoError:
// line 9: function init.0 does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 9: func init must have a body

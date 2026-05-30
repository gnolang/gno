// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type a struct {
  a int
}

func main() {
  av := a{};
  _ = *a(av); // ERROR "invalid indirect|expected pointer|cannot indirect"
}

// GnoError:
// line 15: invalid operation: cannot indirect av<VPBlock(1,0)> (variable of type gno.land/p/filetest/p.a)

// GoTypeCheckError:
// line 15: invalid operation: cannot indirect a(av) (value of struct type a)

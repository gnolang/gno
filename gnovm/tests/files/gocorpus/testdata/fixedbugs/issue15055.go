// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	type name string
	_ = []byte("abc", "def", 12)    // ERROR "too many arguments (to conversion to \[\]byte: \(\[\]byte\)\(.abc., .def., 12\))?"
	_ = string("a", "b", nil)       // ERROR "too many arguments (to conversion to string: string\(.a., .b., nil\))?"
	_ = []byte()                    // ERROR "missing argument (to conversion to \[\]byte: \(\[\]byte\)\(\))?"
	_ = string()                    // ERROR "missing argument (to conversion to string: string\(\))?"
	_ = *int()                      // ERROR "missing argument (to conversion to int: int\(\))?"
	_ = (*int)()                    // ERROR "missing argument (to conversion to \*int: \(\*int\)\(\))?"
	_ = name("a", 1, 3.3)           // ERROR "too many arguments (to conversion to name: name\(.a., 1, 3.3\))?"
	_ = map[string]string(nil, nil) // ERROR "too many arguments (to conversion to map\[string\]string: \(map\[string\]string\)\(nil, nil\))?"
}

// GnoError:
// line 11: type conversion requires single argument
// line 12: type conversion requires single argument
// line 13: type conversion requires single argument
// line 14: type conversion requires single argument
// line 15: type conversion requires single argument
// line 16: type conversion requires single argument
// line 17: type conversion requires single argument
// line 18: type conversion requires single argument

// GoTypeCheckError:
// line 11: too many arguments in conversion to []byte
// line 12: too many arguments in conversion to string
// line 13: missing argument in conversion to []byte
// line 14: missing argument in conversion to string
// line 15: missing argument in conversion to int
// line 16: missing argument in conversion to *int
// line 17: too many arguments in conversion to name
// line 18: too many arguments in conversion to map[string]string

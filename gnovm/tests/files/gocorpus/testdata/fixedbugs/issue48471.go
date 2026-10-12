// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type I interface{ M(int) }

type T struct{}

type T2 struct{}

func (*T2) m(int)

type T3 struct{}

func (*T3) M(string) {}

type T4 struct{}

func (*T4) M(int)

type T5 struct{}

func (T5) m(int) {}

type T6 struct{}

func (T6) m(int) string { return "" }

func f(I)

func g() {
	f(new(T)) // ERROR "cannot use new\(T\) \(.*type \*T\) as I value in argument to f: \*T does not implement I \(missing method M\)"

	var i I
	i = new(T)    // ERROR "cannot use new\(T\) \(.*type \*T\) as I value in assignment: \*T does not implement I \(missing method M\)"
	i = I(new(T)) // ERROR "cannot convert new\(T\) \(.*type \*T\) to type I: \*T does not implement I \(missing method M\)"
	i = new(T2)   // ERROR "cannot use new\(T2\) \(.*type \*T2\) as I value in assignment: \*T2 does not implement I \(missing method M\)\n\t\thave m\(int\)\n\t\twant M\(int\)"

	i = new(T3) // ERROR "cannot use new\(T3\) \(.*type \*T3\) as I value in assignment: \*T3 does not implement I \(wrong type for method M\)\n\t\thave M\(string\)\n\t\twant M\(int\)"

	i = T4{}   // ERROR "cannot use T4\{\} \(.*type T4\) as I value in assignment: T4 does not implement I \(method M has pointer receiver\)"
	i = new(I) // ERROR "cannot use new\(I\) \(.*type \*I\) as I value in assignment: \*I does not implement I \(type \*I is pointer to interface, not interface\)"

	_ = i.(*T2) // ERROR "impossible type assertion: i.\(\*T2\)\n\t\*T2 does not implement I \(missing method M\)\n\t\thave m\(int\)\n\t\twant M\(int\)"
	_ = i.(*T3) // ERROR "impossible type assertion: i.\(\*T3\)\n\t\*T3 does not implement I \(wrong type for method M\)\n\t\thave M\(string\)\n\t\twant M\(int\)"
	_ = i.(T5)  // ERROR ""impossible type assertion: i.\(T5\)\n\tT5 does not implement I \(missing method M\)\n\t\thave m\(int\)\n\t\twant M\(int\)"
	_ = i.(T6)  // ERROR "impossible type assertion: i.\(T6\)\n\tT6 does not implement I \(missing method M\)\n\t\thave m\(int\) string\n\t\twant M\(int\)"

	var t *T4
	t = i // ERROR "cannot use i \(variable of interface type I\) as \*T4 value in assignment: need type assertion"
	_ = t
}

// GnoError:
// line 33: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 39: *gno.land/p/filetest/p.T does not implement gno.land/p/filetest/p.I (missing method M)
// line 40: *gno.land/p/filetest/p.T does not implement gno.land/p/filetest/p.I (missing method M)
// line 41: *gno.land/p/filetest/p.T2 does not implement gno.land/p/filetest/p.I (missing method M)
// line 43: *gno.land/p/filetest/p.T3 does not implement gno.land/p/filetest/p.I (wrong type for method M)
// line 45: gno.land/p/filetest/p.T4 does not implement gno.land/p/filetest/p.I (method M has pointer receiver)
// line 46: should not happen
// line 54: cannot use interface {M func(int)} as *gno.land/p/filetest/p.T4

// GoTypeCheckError:
// line 36
// line 39
// line 40
// line 41
// line 43
// line 45
// line 46
// line 48
// line 49
// line 50
// line 51
// line 54

// GoTypesOnly:
// line 36: gc expects an error (no pattern)
// line 48: gc expects an error (no pattern)
// line 49: gc expects an error (no pattern)
// line 50: gc expects: 
// line 51: gc expects an error (no pattern)

// GnoOverStrictError:
// line 33: function f does not have a body but is not natively defined (did you build after pulling from the repository?)

// Verdict: KnownIssue errorcheck: 1 over-strict line(s), 0 uncaught marker(s)

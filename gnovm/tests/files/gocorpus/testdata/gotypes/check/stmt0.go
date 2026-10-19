// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// statements

package stmt0

func assignments0() (int, int) {
	var a, b, c int
	var ch chan int
	f0 := func() {}
	f1 := func() int { return 1 }
	f2 := func() (int, int) { return 1, 2 }
	f3 := func() (int, int, int) { return 1, 2, 3 }

	a, b, c = 1, 2, 3
	a, b, c = 1 /* ERROR "assignment mismatch: 3 variables but 2 values" */ , 2
	a, b, c = 1 /* ERROR "assignment mismatch: 3 variables but 4 values" */ , 2, 3, 4
	_, _, _ = a, b, c

	a = f0 /* ERROR "used as value" */ ()
	a = f1()
	a = f2 /* ERROR "assignment mismatch: 1 variable but f2 returns 2 values" */ ()
	a, b = f2()
	a, b, c = f2 /* ERROR "assignment mismatch: 3 variables but f2 returns 2 values" */ ()
	a, b, c = f3()
	a, b = f3 /* ERROR "assignment mismatch: 2 variables but f3 returns 3 values" */ ()

	a, b, c = <- /* ERROR "assignment mismatch: 3 variables but 1 value" */ ch

	return /* ERROR "not enough return values\n\thave ()\n\twant (int, int)" */
	return 1 /* ERROR "not enough return values\n\thave (number)\n\twant (int, int)" */
	return 1, 2
	return 1, 2, 3 /* ERROR "too many return values\n\thave (number, number, number)\n\twant (int, int)" */
}

func assignments1() {
	b, i, f, c, s := false, 1, 1.0, 1i, "foo"
	b = i /* ERRORx `cannot use .* in assignment` */
	i = f /* ERRORx `cannot use .* in assignment` */
	f = c /* ERRORx `cannot use .* in assignment` */
	c = s /* ERRORx `cannot use .* in assignment` */
	s = b /* ERRORx `cannot use .* in assignment` */

	v0, v1, v2 := 1 /* ERROR "assignment mismatch" */ , 2, 3, 4
	_, _, _ = v0, v1, v2

	b = true

	i += 1
	i /* ERROR "mismatched types int and untyped string" */+= "foo"

	f -= 1
	f /= 0
	f = float32(0)/0 /* ERROR "division by zero" */
	f /* ERROR "mismatched types float64 and untyped string" */-= "foo"

	c *= 1
	c /= 0

	s += "bar"
	s /* ERROR "mismatched types string and untyped int" */+= 1

	var u64 uint64
	u64 += 1<<u64

	undefined /* ERROR "undefined" */ = 991

	// test cases for issue 5800
	var (
		_ int = nil /* ERROR "cannot use nil as int value in variable declaration" */
		_ [10]int = nil /* ERROR "cannot use nil as [10]int value in variable declaration" */
		_ []byte = nil
		_ struct{} = nil /* ERROR "cannot use nil as struct{} value in variable declaration" */
		_ func() = nil
		_ map[int]string = nil
		_ chan int = nil
	)

	// test cases for issue 5500
	_ = func() (int, bool) {
		var m map[int]int
		return m /* ERROR "not enough return values" */ [0]
	}

	g := func(int, bool){}
	var m map[int]int
	g(m[0]) /* ERROR "not enough arguments" */

	// assignments to _
	_ = nil /* ERROR "use of untyped nil" */
	_ = 1  << /* ERROR "constant shift overflow" */ 1000
	(_) = 0
}

func assignments2() {
	type mybool bool
	var m map[string][]bool
	var s []bool
	var b bool
	var d mybool
	_ = s
	_ = b
	_ = d

	// assignments to map index expressions are ok
	s, b = m["foo"]
	_, d = m["bar"]
	m["foo"] = nil
	m["foo"] = nil /* ERROR "assignment mismatch: 1 variable but 2 values" */ , false
	_ = append(m["foo"])
	_ = append(m["foo"], true)

	var c chan int
	_, b = <-c
	_, d = <-c
	<- /* ERROR "cannot assign" */ c = 0
	<-c = 0 /* ERROR "assignment mismatch: 1 variable but 2 values" */ , false

	var x interface{}
	_, b = x.(int)
	x /* ERROR "cannot assign" */ .(int) = 0
	x.(int) = 0 /* ERROR "assignment mismatch: 1 variable but 2 values" */ , false

	assignments2 /* ERROR "used as value" */ () = nil
	int /* ERROR "not an expression" */ = 0
}

func issue6487() {
	type S struct{x int}
	_ = &S /* ERROR "cannot take address" */ {}.x
	_ = &( /* ERROR "cannot take address" */ S{}.x)
	_ = (&S{}).x
	S /* ERROR "cannot assign" */ {}.x = 0
	(&S{}).x = 0

	type M map[string]S
	var m M
	m /* ERROR "cannot assign to struct field" */ ["foo"].x = 0
	_ = &( /* ERROR "cannot take address" */ m["foo"].x)
	_ = &m /* ERROR "cannot take address" */ ["foo"].x
}

func issue6766a() {
	a, a /* ERROR "a repeated on left side of :=" */ := 1, 2
	_ = a
	a, b, b /* ERROR "b repeated on left side of :=" */ := 1, 2, 3
	_ = b
	c, c /* ERROR "c repeated on left side of :=" */, b := 1, 2, 3
	_ = c
	a, b := /* ERROR "no new variables" */ 1, 2
}

func shortVarDecls1() {
	const c = 0
	type d int
	a, b, c /* ERROR "cannot assign" */ , d /* ERROR "cannot assign" */  := 1, "zwei", 3.0, 4
	var _ int = a // a is of type int
	var _ string = b // b is of type string
}

func incdecs() {
	const c = 3.14
	c /* ERROR "cannot assign" */ ++
	s := "foo"
	s /* ERROR "invalid operation" */ --
	3.14 /* ERROR "cannot assign" */ ++
	var (
		x int
		y float32
		z complex128
	)
	x++
	y--
	z++
}

func sends() {
	var ch chan int
	var rch <-chan int
	var x int
	x <- /* ERROR "cannot send" */ x
	rch <- /* ERROR "cannot send" */ x
	ch <- "foo" /* ERRORx `cannot use .* in send` */
	ch <- x
}

func selects() {
	select {}
	var (
		ch chan int
		sc chan <- bool
	)
	select {
	case <-ch:
	case (<-ch):
	case t := <-ch:
		_ = t
	case t := (<-ch):
		_ = t
	case t, ok := <-ch:
		_, _ = t, ok
	case t, ok := (<-ch):
		_, _ = t, ok
	case <-sc /* ERROR "cannot receive from send-only channel" */ :
	}
	select {
	default:
	default /* ERROR "multiple defaults" */ :
	}
	select {
	case a, b := <-ch:
		_, b = a, b
	case x /* ERROR "send or receive" */ :
	case a /* ERROR "send or receive" */ := ch:
	}

	// test for issue 9570: ch2 in second case falsely resolved to
	// ch2 declared in body of first case
	ch1 := make(chan int)
	ch2 := make(chan int)
	select {
	case <-ch1:
		var ch2 /* ERROR "declared and not used: ch2" */ chan bool
	case i := <-ch2:
		print(i + 1)
	}
}

func gos() {
	go 1; /* ERROR "must be function call" */
	go int /* ERROR "go requires function call, not conversion" */ (0)
	go ( /* ERROR "expression in go must not be parenthesized" */ gos())
	go gos()
	var c chan int
	go close(c)
	go len /* ERROR "go discards result" */ (c)
}

func defers() {
	defer 1; /* ERROR "must be function call" */
	defer int /* ERROR "defer requires function call, not conversion" */ (0)
	defer ( /* ERROR "expression in defer must not be parenthesized" */ defers())
	defer defers()
	var c chan int
	defer close(c)
	defer len /* ERROR "defer discards result" */ (c)
}

func breaks() {
	var x, y int

	break /* ERROR "break" */
	{
		break /* ERROR "break" */
	}
	if x < y {
		break /* ERROR "break" */
	}

	switch x {
	case 0:
		break
	case 1:
		if x == y {
			break
		}
	default:
		break
		break
	}

	var z interface{}
	switch z.(type) {
	case int:
		break
	}

	for {
		break
	}

	var a []int
	for _ = range a {
		break
	}

	for {
		if x == y {
			break
		}
	}

	var ch chan int
	select {
	case <-ch:
		break
	}

	select {
	case <-ch:
		if x == y {
			break
		}
	default:
		break
	}
}

func continues() {
	var x, y int

	continue /* ERROR "continue" */
	{
		continue /* ERROR "continue" */
	}

	if x < y {
		continue /* ERROR "continue" */
	}

	switch x {
	case 0:
		continue /* ERROR "continue" */
	}

	var z interface{}
	switch z.(type) {
	case int:
		continue /* ERROR "continue" */
	}

	var ch chan int
	select {
	case <-ch:
		continue /* ERROR "continue" */
	}

	for i := 0; i < 10; i++ {
		continue
		if x < y {
			continue
			break
		}
		switch x {
		case y:
			continue
		default:
			break
		}
		select {
		case <-ch:
			continue
		}
	}

	var a []int
	for _ = range a {
		continue
		if x < y {
			continue
			break
		}
		switch x {
		case y:
			continue
		default:
			break
		}
		select {
		case <-ch:
			continue
		}
	}
}

func returns0() {
	return
	return 0 /* ERROR "too many return values" */
}

func returns1(x float64) (int, *float64) {
	return 0, &x
	return /* ERROR "not enough return values" */
	return "foo" /* ERRORx `cannot .* in return statement` */, x /* ERRORx `cannot use .* in return statement` */
	return 0, &x, 1 /* ERROR "too many return values" */
}

func returns2() (a, b int) {
	return
	return 1, "foo" /* ERRORx `cannot use .* in return statement` */
	return 1, 2, 3 /* ERROR "too many return values" */
	{
		type a int
		return 1, 2
		return /* ERROR "result parameter a not in scope at return" */
	}
}

func returns3() (_ int) {
	return
	{
		var _ int // blank (_) identifiers never shadow since they are in no scope
		return
	}
}

func switches0() {
	var x int

	switch x {
	}

	switch x {
	default:
	default /* ERROR "multiple defaults" */ :
	}

	switch {
	case 1  /* ERROR "cannot convert" */ :
	}

	true := "false"
	_ = true
	// A tagless switch is equivalent to the bool
        // constant true, not the identifier 'true'.
	switch {
	case "false" /* ERROR "cannot convert" */:
	}

	switch int32(x) {
	case 1, 2:
	case x /* ERROR "invalid case x in switch on int32(x) (mismatched types int and int32)" */ :
	}

	switch x {
	case 1 /* ERROR "overflows" */ << 100:
	}

	switch x {
	case 1:
	case 1 /* ERROR "duplicate case" */ :
	case ( /* ERROR "duplicate case" */ 1):
	case 2, 3, 4:
	case 5, 1 /* ERROR "duplicate case" */ :
	}

	switch uint64(x) {
	case 1<<64 - 1:
	case 1 /* ERROR "duplicate case" */ <<64 - 1:
	case 2, 3, 4:
	case 5, 1 /* ERROR "duplicate case" */ <<64 - 1:
	}

	var y32 float32
	switch y32 {
	case 1.1:
	case 11/10: // integer division!
	case 11. /* ERROR "duplicate case" */ /10:
	case 2, 3.0, 4.1:
	case 5.2, 1.10 /* ERROR "duplicate case" */ :
	}

	var y64 float64
	switch y64 {
	case 1.1:
	case 11/10: // integer division!
	case 11. /* ERROR "duplicate case" */ /10:
	case 2, 3.0, 4.1:
	case 5.2, 1.10 /* ERROR "duplicate case" */ :
	}

	var s string
	switch s {
	case "foo":
	case "foo" /* ERROR "duplicate case" */ :
	case "f" /* ERROR "duplicate case" */ + "oo":
	case "abc", "def", "ghi":
	case "jkl", "foo" /* ERROR "duplicate case" */ :
	}

	type T int
	type F float64
	type S string
	type B bool
	var i interface{}
	switch i {
	case nil:
	case nil: // no duplicate detection
	case (*int)(nil):
	case (*int)(nil): // do duplicate detection
	case 1:
	case byte(1):
	case int /* ERROR "duplicate case" */ (1):
	case T(1):
	case 1.0:
	case F(1.0):
	case F /* ERROR "duplicate case" */ (1.0):
	case "hello":
	case S("hello"):
	case S /* ERROR "duplicate case" */ ("hello"):
	case 1==1, B(false):
	case false, B(2==2):
	}

	// switch on array
	var a [3]int
	switch a {
	case [3]int{1, 2, 3}:
	case [3]int{1, 2, 3}: // no duplicate detection
	case [ /* ERROR "mismatched types" */ 4]int{4, 5, 6}:
	}

	// switch on channel
	var c1, c2 chan int
	switch c1 {
	case nil:
	case c1:
	case c2:
	case c1, c2: // no duplicate detection
	}
}

func switches1() {
	fallthrough /* ERROR "fallthrough statement out of place" */

	var x int
	switch x {
	case 0:
		fallthrough /* ERROR "fallthrough statement out of place" */
		break
	case 1:
		fallthrough
	case 2:
		fallthrough; ; ; // trailing empty statements are ok
	case 3:
	default:
		fallthrough; ;
	case 4:
		fallthrough /* ERROR "cannot fallthrough final case in switch" */
	}

	var y interface{}
	switch y.(type) {
	case int:
		fallthrough /* ERROR "cannot fallthrough in type switch" */ ; ; ;
	default:
	}

	switch x {
	case 0:
		if x == 0 {
			fallthrough /* ERROR "fallthrough statement out of place" */
		}
	}

	switch x {
	case 0:
		goto L1
		L1: fallthrough; ;
	case 1:
		goto L2
		goto L3
		goto L4
		L2: L3: L4: fallthrough
	default:
	}

	switch x {
	case 0:
		goto L5
		L5: fallthrough
	default:
		goto L6
		goto L7
		goto L8
		L6: L7: L8: fallthrough /* ERROR "cannot fallthrough final case in switch" */
	}

	switch x {
	case 0:
		fallthrough; ;
	case 1:
		{
			fallthrough /* ERROR "fallthrough statement out of place" */
		}
	case 2:
		fallthrough
	case 3:
		fallthrough /* ERROR "fallthrough statement out of place" */
		{ /* empty block is not an empty statement */ }; ;
	default:
		fallthrough /* ERROR "cannot fallthrough final case in switch" */
	}

	switch x {
	case 0:
		{
			fallthrough /* ERROR "fallthrough statement out of place" */
		}
	}
}

func switches2() {
	// untyped nil is not permitted as switch expression
	switch nil /* ERROR "use of untyped nil" */ {
	case 1, 2, "foo": // don't report additional errors here
	}

	// untyped constants are converted to default types
	switch 1<<63-1 {
	}
	switch 1 /* ERRORx `cannot use .* as int value.*\(overflows\)` */ << 63 {
	}
	var x int
	switch 1.0 {
	case 1.0, 2.0, x /* ERROR "mismatched types int and float64" */ :
	}
	switch x {
	case 1.0:
	}

	// untyped bools become of type bool
	type B bool
	var b B = true
	switch x == x {
	case b /* ERROR "mismatched types B and bool" */ :
	}
	switch {
	case b /* ERROR "mismatched types B and bool" */ :
	}
}

func issue11667() {
	switch 9223372036854775808 /* ERRORx `cannot use .* as int value.*\(overflows\)` */ {
	}
	switch 9223372036854775808 /* ERRORx `cannot use .* as int value.*\(overflows\)` */ {
	case 9223372036854775808:
	}
	var x int
	switch x {
	case 9223372036854775808 /* ERROR "overflows int" */ :
	}
	var y float64
	switch y {
	case 9223372036854775808:
	}
}

func issue11687() {
	f := func() (_, _ int) { return }
	switch f /* ERROR "multiple-value f" */ () {
	}
	var x int
	switch f /* ERROR "multiple-value f" */ () {
	case x:
	}
	switch x {
	case f /* ERROR "multiple-value f" */ ():
	}
}

type I interface {
	m()
}

type I2 interface {
	m(int)
}

type T struct{}
type T1 struct{}
type T2 struct{}

func (T) m() {}
func (T2) m(int) {}

func typeswitches() {
	var i int
	var x interface{}

	switch x.(type) {}
	switch (x /* ERROR "outside type switch" */ .(type)) {}

	switch x.(type) {
	default:
	default /* ERROR "multiple defaults" */ :
	}

	switch x /* ERROR "declared and not used" */ := x.(type) {}
	switch _ /* ERROR "no new variable on left side of :=" */ := x.(type) {}

	switch x := x.(type) {
	case int:
		var y int = x
		_ = y
	}

	switch x /* ERROR "x declared and not used" */ := i /* ERROR "not an interface" */ .(type) {}

	switch t := x.(type) {
	case nil:
		var v bool = t /* ERRORx `cannot use .* in variable declaration` */
		_ = v
	case int:
		var v int = t
		_ = v
	case float32, complex64:
		var v float32 = t /* ERRORx `cannot use .* in variable declaration` */
		_ = v
	default:
		var v float32 = t /* ERRORx `cannot use .* in variable declaration` */
		_ = v
	}

	var t I
	switch t.(type) {
	case T:
	case T1 /* ERROR "missing method m" */ :
	case T2 /* ERROR "wrong type for method m" */ :
	case I2 /* STRICT "wrong type for method m" */ : // only an error in strict mode (issue 8561)
	}


	{
		x := 1
		v := 2
		switch v /* ERROR "v (variable of type int) is not an interface" */ .(type) {
		case int:
			println(x)
			println(x / 0 /* ERROR "invalid operation: division by zero" */)
		case 1 /* ERROR "1 is not a type" */:
		}
	}
}

// Test that each case clause uses the correct type of the variable
// declared by the type switch (issue 5504).
func typeswitch0() {
	switch y := interface{}(nil).(type) {
	case int:
		func() int { return y + 0 }()
	case float32:
		func() float32 { return y }()
	}
}

// Test correct scope setup.
// (no redeclaration errors expected in the type switch)
func typeswitch1() {
	var t I
	switch t := t; t := t.(type) {
	case nil:
		var _ I = t
	case T:
		var _ T = t
	default:
		var _ I = t
	}
}

// Test correct typeswitch against interface types.
type A interface { a() }
type B interface { b() }
type C interface { a(int) }

func typeswitch2() {
	switch A(nil).(type) {
	case A:
	case B:
	case C /* STRICT "cannot have dynamic type" */: // only an error in strict mode (issue 8561)
	}
}

func typeswitch3(x interface{}) {
	switch x.(type) {
	case int:
	case float64:
	case int /* ERROR "duplicate case" */ :
	}

	switch x.(type) {
	case nil:
	case int:
	case nil /* ERROR "duplicate case" */ , nil /* ERROR "duplicate case" */ :
	}

	type F func(int)
	switch x.(type) {
	case nil:
	case int, func(int):
	case float32, func /* ERROR "duplicate case" */ (x int):
	case F:
	}
}

func fors1() {
	for {}
	var i string
	_ = i
	for i := 0; i < 10; i++ {}
	for i := 0; i < 10; j /* ERROR "cannot declare" */ := 0 {}
}

func rangeloops1() {
	var (
		a [10]float32
		b []string
		p *[10]complex128
		pp **[10]complex128
		s string
		m map[int]bool
		c chan int
		sc chan<- int
		rc <-chan int
		xs struct{}
	)

	for range xs /* ERROR "cannot range over" */ {}
	for _ = range xs /* ERROR "cannot range over" */ {}
	for i := range xs /* ERROR "cannot range over" */ { _ = i }

	for range a {}
	for i := range a {
		var ii int
		ii = i
		_ = ii
	}
	for i, x := range a {
		var ii int
		ii = i
		_ = ii
		var xx float64
		xx = x /* ERRORx `cannot use .* in assignment` */
		_ = xx
	}
	var ii int
	var xx float32
	for ii, xx = range a {}
	_, _ = ii, xx

	for range b {}
	for i := range b {
		var ii int
		ii = i
		_ = ii
	}
	for i, x := range b {
		var ii int
		ii = i
		_ = ii
		var xx string
		xx = x
		_ = xx
	}

	for range s {}
	for i := range s {
		var ii int
		ii = i
		_ = ii
	}
	for i, x := range s {
		var ii int
		ii = i
		_ = ii
		var xx rune
		xx = x
		_ = xx
	}

	for range p {}
	for _, x := range p {
		var xx complex128
		xx = x
		_ = xx
	}

	for range pp /* ERROR "cannot range over" */ {}
	for _, x := range pp /* ERROR "cannot range over" */ {}

	for range m {}
	for k := range m {
		var kk int32
		kk = k /* ERRORx `cannot use .* in assignment` */
		_ = kk
	}
	for k, v := range m {
		var kk int
		kk = k
		_ = kk
		if v {}
	}

	for range c {}
	for _, _ /* ERROR "only one iteration variable" */ = range c {}
	for e := range c {
		var ee int
		ee = e
		_ = ee
	}
	for _ = range sc /* ERROR "cannot range over" */ {}
	for _ = range rc {}

	// constant strings
	const cs = "foo"
	for range cs {}
	for range "" {}
	for i, x := range cs { _, _ = i, x }
	for i, x := range "" {
		var ii int
		ii = i
		_ = ii
		var xx rune
		xx = x
		_ = xx
	}
}

func rangeloops2() {
	type I int
	type R rune

	var a [10]int
	var i I
	_ = i
	for i /* ERRORx `cannot use .* in assignment` */ = range a {}
	for i /* ERRORx `cannot use .* in assignment` */ = range &a {}
	for i /* ERRORx `cannot use .* in assignment` */ = range a[:] {}

	var s string
	var r R
	_ = r
	for i /* ERRORx `cannot use .* in assignment` */ = range s {}
	for i /* ERRORx `cannot use .* in assignment` */ = range "foo" {}
	for _, r /* ERRORx `cannot use .* in assignment` */ = range s {}
	for _, r /* ERRORx `cannot use .* in assignment` */ = range "foo" {}
}

func issue6766b() {
	for _ := /* ERROR "no new variables" */ range "" {}
	for a, a /* ERROR "redeclared" */ := range "" { _ = a }
	var a int
	_ = a
	for a, a /* ERROR "redeclared" */ := range []int{1, 2, 3} { _ = a }
}

// Test that despite errors in the range clause,
// the loop body is still type-checked (and thus
// errors reported).
func issue10148() {
	for y /* ERROR "declared and not used" */ := range "" {
		_ = "" /* ERROR "mismatched types untyped string and untyped int" */ + 1
	}
	for range 1.5 /* ERROR "cannot range over 1.5 (untyped float constant)" */ {
		_ = "" /* ERROR "mismatched types untyped string and untyped int" */ + 1
	}
	for y := range 1.5 /* ERROR "cannot range over 1.5 (untyped float constant)" */ {
		_ = "" /* ERROR "mismatched types untyped string and untyped int" */ + 1
	}
}

func labels0() {
	goto L0
	goto L1
	L0:
	L1:
	L1 /* ERROR "already declared" */ :
	if true {
		goto L2
		L2:
		L0 /* ERROR "already declared" */ :
	}
	_ = func() {
		goto L0
		goto L1
		goto L2
		L0:
		L1:
		L2:
	}
}

func expression_statements(ch chan int) {
	expression_statements(ch)
	<-ch
	println()

	0 /* ERROR "not used" */
	1 /* ERROR "not used" */ +2
	cap /* ERROR "not used" */ (ch)
	println /* ERROR "must be called" */
}

// GnoError:
// line 152: no new variables on left side of :=
// line 232: expression in go must be function call
// line 234: expression in go must not be parenthesized
// line 242: expression in defer must be function call
// line 244: expression in defer must not be parenthesized
// line 692: no new variables on left side of :=
// line 942: no new variables on left side of :=
// line 969: L1 redeclared in this block
// 	previous declaration at stmt0.go:969:2
// line 973: L0 redeclared in this block
// 	previous declaration at stmt0.go:968:2

// ParseBlocked:
// line 18: gc expects: assignment mismatch: 3 variables but 2 values
// line 19: gc expects: assignment mismatch: 3 variables but 4 values
// line 22: gc expects: used as value
// line 24: gc expects: assignment mismatch: 1 variable but f2 returns 2 values
// line 26: gc expects: assignment mismatch: 3 variables but f2 returns 2 values
// line 28: gc expects: assignment mismatch: 2 variables but f3 returns 3 values
// line 30: gc expects: assignment mismatch: 3 variables but 1 value
// line 32: gc expects: not enough return values
// 	have ()
// 	want (int, int)
// line 33: gc expects: not enough return values
// 	have (number)
// 	want (int, int)
// line 35: gc expects: too many return values
// 	have (number, number, number)
// 	want (int, int)
// line 40: gc expects: cannot use .* in assignment
// line 41: gc expects: cannot use .* in assignment
// line 42: gc expects: cannot use .* in assignment
// line 43: gc expects: cannot use .* in assignment
// line 44: gc expects: cannot use .* in assignment
// line 46: gc expects: assignment mismatch
// line 52: gc expects: mismatched types int and untyped string
// line 56: gc expects: division by zero
// line 57: gc expects: mismatched types float64 and untyped string
// line 63: gc expects: mismatched types string and untyped int
// line 68: gc expects: undefined
// line 72: gc expects: cannot use nil as int value in variable declaration
// line 73: gc expects: cannot use nil as [10]int value in variable declaration
// line 75: gc expects: cannot use nil as struct{} value in variable declaration
// line 84: gc expects: not enough return values
// line 89: gc expects: not enough arguments
// line 92: gc expects: use of untyped nil
// line 93: gc expects: constant shift overflow
// line 111: gc expects: assignment mismatch: 1 variable but 2 values
// line 118: gc expects: cannot assign
// line 119: gc expects: assignment mismatch: 1 variable but 2 values
// line 123: gc expects: cannot assign
// line 124: gc expects: assignment mismatch: 1 variable but 2 values
// line 126: gc expects: used as value
// line 127: gc expects: not an expression
// line 132: gc expects: cannot take address
// line 133: gc expects: cannot take address
// line 135: gc expects: cannot assign
// line 140: gc expects: cannot assign to struct field
// line 141: gc expects: cannot take address
// line 142: gc expects: cannot take address
// line 146: gc expects: a repeated on left side of :=
// line 148: gc expects: b repeated on left side of :=
// line 150: gc expects: c repeated on left side of :=
// line 158: gc expects: cannot assign | cannot assign
// line 165: gc expects: cannot assign
// line 167: gc expects: invalid operation
// line 168: gc expects: cannot assign
// line 183: gc expects: cannot send
// line 184: gc expects: cannot send
// line 185: gc expects: cannot use .* in send
// line 206: gc expects: cannot receive from send-only channel
// line 210: gc expects: multiple defaults
// line 215: gc expects: send or receive
// line 216: gc expects: send or receive
// line 225: gc expects: declared and not used: ch2
// line 233: gc expects: go requires function call, not conversion
// line 238: gc expects: go discards result
// line 243: gc expects: defer requires function call, not conversion
// line 248: gc expects: defer discards result
// line 254: gc expects: break
// line 256: gc expects: break
// line 259: gc expects: break
// line 314: gc expects: continue
// line 316: gc expects: continue
// line 320: gc expects: continue
// line 325: gc expects: continue
// line 331: gc expects: continue
// line 337: gc expects: continue
// line 380: gc expects: too many return values
// line 385: gc expects: not enough return values
// line 386: gc expects: cannot .* in return statement | cannot use .* in return statement
// line 387: gc expects: too many return values
// line 392: gc expects: cannot use .* in return statement
// line 393: gc expects: too many return values
// line 397: gc expects: result parameter a not in scope at return
// line 417: gc expects: multiple defaults
// line 421: gc expects: cannot convert
// line 429: gc expects: cannot convert
// line 434: gc expects: invalid case x in switch on int32(x) (mismatched types int and int32)
// line 438: gc expects: overflows
// line 443: gc expects: duplicate case
// line 444: gc expects: duplicate case
// line 446: gc expects: duplicate case
// line 451: gc expects: duplicate case
// line 453: gc expects: duplicate case
// line 460: gc expects: duplicate case
// line 462: gc expects: duplicate case
// line 469: gc expects: duplicate case
// line 471: gc expects: duplicate case
// line 477: gc expects: duplicate case
// line 478: gc expects: duplicate case
// line 480: gc expects: duplicate case
// line 495: gc expects: duplicate case
// line 499: gc expects: duplicate case
// line 502: gc expects: duplicate case
// line 512: gc expects: mismatched types
// line 526: gc expects: fallthrough statement out of place
// line 531: gc expects: fallthrough statement out of place
// line 541: gc expects: cannot fallthrough final case in switch
// line 547: gc expects: cannot fallthrough in type switch
// line 554: gc expects: fallthrough statement out of place
// line 578: gc expects: cannot fallthrough final case in switch
// line 586: gc expects: fallthrough statement out of place
// line 591: gc expects: fallthrough statement out of place
// line 594: gc expects: cannot fallthrough final case in switch
// line 600: gc expects: fallthrough statement out of place
// line 607: gc expects: use of untyped nil
// line 614: gc expects: cannot use .* as int value.*\(overflows\)
// line 618: gc expects: mismatched types int and float64
// line 628: gc expects: mismatched types B and bool
// line 631: gc expects: mismatched types B and bool
// line 636: gc expects: cannot use .* as int value.*\(overflows\)
// line 638: gc expects: cannot use .* as int value.*\(overflows\)
// line 643: gc expects: overflows int
// line 653: gc expects: multiple-value f
// line 656: gc expects: multiple-value f
// line 660: gc expects: multiple-value f
// line 684: gc expects: outside type switch
// line 688: gc expects: multiple defaults
// line 691: gc expects: declared and not used
// line 700: gc expects: x declared and not used | not an interface
// line 704: gc expects: cannot use .* in variable declaration
// line 710: gc expects: cannot use .* in variable declaration
// line 713: gc expects: cannot use .* in variable declaration
// line 720: gc expects: missing method m
// line 721: gc expects: wrong type for method m
// line 729: gc expects: v (variable of type int) is not an interface
// line 732: gc expects: invalid operation: division by zero
// line 733: gc expects: 1 is not a type
// line 780: gc expects: duplicate case
// line 786: gc expects: duplicate case | duplicate case
// line 793: gc expects: duplicate case
// line 803: gc expects: cannot declare
// line 820: gc expects: cannot range over
// line 821: gc expects: cannot range over
// line 822: gc expects: cannot range over
// line 835: gc expects: cannot use .* in assignment
// line 880: gc expects: cannot range over
// line 881: gc expects: cannot range over
// line 886: gc expects: cannot use .* in assignment
// line 897: gc expects: only one iteration variable
// line 903: gc expects: cannot range over
// line 928: gc expects: cannot use .* in assignment
// line 929: gc expects: cannot use .* in assignment
// line 930: gc expects: cannot use .* in assignment
// line 935: gc expects: cannot use .* in assignment
// line 936: gc expects: cannot use .* in assignment
// line 937: gc expects: cannot use .* in assignment
// line 938: gc expects: cannot use .* in assignment
// line 943: gc expects: redeclared
// line 946: gc expects: redeclared
// line 953: gc expects: declared and not used
// line 954: gc expects: mismatched types untyped string and untyped int
// line 956: gc expects: cannot range over 1.5 (untyped float constant)
// line 957: gc expects: mismatched types untyped string and untyped int
// line 959: gc expects: cannot range over 1.5 (untyped float constant)
// line 960: gc expects: mismatched types untyped string and untyped int
// line 990: gc expects: not used
// line 991: gc expects: not used
// line 992: gc expects: not used
// line 993: gc expects: must be called

// GnoWordingDiffers:
// line 692: gc expects: no new variable on left side of :=; gno: no new variables on left side of :=
// line 969: gc expects: already declared; gno: L1 redeclared in this block
// 	previous declaration at stmt0.go:969:2
// line 973: gc expects: already declared; gno: L0 redeclared in this block
// 	previous declaration at stmt0.go:968:2

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 18, 19, 22, 24, 26, 28, 30, 32, 33, 35, 40, 41, 42, 43, 44, 46, 52, 56, 57, 63, 68, 72, 73, 75, 84, 89, 92, 93, 111, 118, 119, 123, 124, 126, 127, 132, 133, 135, 140, 141, 142, 146, 148, 150, 158, 165, 167, 168, 183, 184, 185, 206, 210, 215, 216, 225, 233, 238, 243, 248, 254, 256, 259, 314, 316, 320, 325, 331, 337, 380, 385, 386, 387, 392, 393, 397, 417, 421, 429, 434, 438, 443, 444, 446, 451, 453, 460, 462, 469, 471, 477, 478, 480, 495, 499, 502, 512, 526, 531, 541, 547, 554, 578, 586, 591, 594, 600, 607, 614, 618, 628, 631, 636, 638, 643, 653, 656, 660, 684, 688, 691, 700, 704, 710, 713, 720, 721, 729, 732, 733, 780, 786, 793, 803, 820, 821, 822, 835, 880, 881, 886, 897, 903, 928, 929, 930, 935, 936, 937, 938, 943, 946, 953, 954, 956, 957, 959, 960, 990, 991, 992, 993

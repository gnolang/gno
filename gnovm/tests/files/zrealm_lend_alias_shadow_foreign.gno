// PKGPATH: gno.land/r/tests/vm/lendaliasshadowfor
package lendaliasshadowfor

// The victim calls this caller's value (its own writer ExName) through a
// captured name; a same-named alias or helper declared later must not change that.

import (
	v "gno.land/r/tests/vm/t76victim"
)

var m int

func show(cur realm, s string, err any) {
	if e, ok := err.(string); ok && len(e) > 90 {
		err = e[:90]
	}
	println(s, v.Get(), m, err)
	v.Reset(cross(cur))
	m = 0
}

func do(cur realm, which, note string, f func(), g func(int)) {
	show(cur, which+" "+note, v.Do(cross(cur), which, f, g))
}

func main(cur realm) {
	p := v.Ptr()
	for _, w := range []string{"CapCtl", "CapParam", "CapLocal", "CapHelper"} {
		do(cur, w, "caller lit writes victim ptr", func() { *p = 50 }, nil)
		do(cur, w, "victim writer by value", v.ExName, nil)
		do(cur, w, "caller lit writes caller var", func() { m++ }, nil)
	}
	do(cur, "CbLibName", "g control", nil, func(x int) { *p = x })
	do(cur, "CapG", "caller lit writes victim ptr", nil, func(x int) { *p = x })
}

// Output:
// CapCtl caller lit writes victim ptr 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CapCtl victim writer by value 0 0 gno.land/r/tests/vm/t76victim's code is read-only here: a function value writes only as a
// CapCtl caller lit writes caller var 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CapParam caller lit writes victim ptr 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CapParam victim writer by value 0 0 gno.land/r/tests/vm/t76victim's code is read-only here: a function value writes only as a
// CapParam caller lit writes caller var 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CapLocal caller lit writes victim ptr 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CapLocal victim writer by value 0 0 gno.land/r/tests/vm/t76victim's code is read-only here: a function value writes only as a
// CapLocal caller lit writes caller var 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CapHelper caller lit writes victim ptr 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CapHelper victim writer by value 0 0 gno.land/r/tests/vm/t76victim's code is read-only here: a function value writes only as a
// CapHelper caller lit writes caller var 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CbLibName g control 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl
// CapG caller lit writes victim ptr 0 0 gno.land/r/tests/vm/lendaliasshadowfor's code was entered without crossing and is read-onl

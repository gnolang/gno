// PKGPATH: gno.land/r/launderrh_conv_val_then_addr
package launderrh_conv_val_then_addr

// CONVERSION VECTOR 9: copy struct VALUE then write through it.
//
// SURPRISE FINDING: Gno's readonly taint is CONSERVATIVE — it travels
// with the value copy. `b := victim.Slice[0]` followed by `b.Field =
// "..."` is BLOCKED, even though `b` is a fresh local variable. The
// readonly-bit is encoded on the value itself, not just on its
// container.
//
// This is over-blocking from a Go semantics perspective (Go would
// allow mutating the local copy), but is safe: an attacker can't
// abuse the local copy to launder writes back to the victim.

import (
	"gno.land/p/demo/tests/launderpkg"
	"gno.land/r/tests/vm/launderrvictim"
)

func try(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			println(label, "blocked: true")
		} else {
			println(label, "blocked: false")
		}
	}()
	fn()
}

func main(cur realm) {
	// 1. Copy value, take address, write local. Should succeed.
	try("1 local copy + addr write", func() {
		b := launderrvictim.GetBareSlice()[0] // value copy
		p := &b
		p.Field = "local-only"
		_ = p
	})

	// 2. Convert (no-op same-type) value, write through addr.
	//    `launderpkg.Bare(...)` form. Should succeed (Case 1 only
	//    fires on foreign-readonly source — value copy is local).
	try("2 same-type conv on copy", func() {
		b := launderrvictim.GetBareSlice()[0]
		b2 := launderpkg.Bare(b)
		p := &b2
		p.Field = "local-conv"
	})

	// 3. Try taking address of an indexed expression (not a copy).
	//    s := getSlice(); p := &s[0]. The pointer aliases victim,
	//    NOT a copy. Write should be blocked.
	try("3 &s[0] write", func() {
		s := launderrvictim.GetBareSlice()
		p := &s[0]
		p.Field = "pwnd-via-elem-addr"
	})

	// 4. Try deref-copy then re-assign. *p2 = *p1 type pattern.
	//    Assignment-to-victim via deref. Should be blocked.
	try("4 *victimPtr = newValue", func() {
		w := launderrvictim.GetWithBarePtr()
		*w.Inner = launderpkg.Bare{Field: "pwnd-deref-assign"}
	})

	// 5. Copy struct, assign back into victim slot via Inner =.
	//    Write side blocks because Inner field lives in victim.
	try("5 victim.Inner = localCopy", func() {
		w := launderrvictim.GetWithBarePtr()
		local := launderpkg.Bare{Field: "fake"}
		w.Inner = &local
	})

	println("--- final reads ---")
	println("ptr:", launderrvictim.ReadBarePtr())
	println("slice[0]:", launderrvictim.ReadBareSlice0())
}

// Output:
// 1 local copy + addr write blocked: true
// 2 same-type conv on copy blocked: true
// 3 &s[0] write blocked: true
// 4 *victimPtr = newValue blocked: true
// 5 victim.Inner = localCopy blocked: true
// --- final reads ---
// ptr: bare-ptr-orig
// slice[0]: slice0

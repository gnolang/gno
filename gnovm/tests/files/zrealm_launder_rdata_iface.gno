// PKGPATH: gno.land/r/launderrh_iface
package launderrh_iface

// Interface-boxing / unboxing variants against /r/-victim with
// embedded /p/-Bare. The question: does N_Readonly taint survive
// being boxed into `any` and then type-asserted back out?

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
	// 1. Box pointer, type-assert back to *Bare, write
	try("1 any.(*Bare).Field write", func() {
		var box any = launderrvictim.GetWithBarePtr().Inner
		t := box.(*launderpkg.Bare)
		t.Field = "pwnd-via-any"
	})

	// 2. Box pointer THROUGH a generic-any-passing local fn,
	//    type-assert in receiver
	try("2 fn(any)→assert→write", func() {
		writeViaAny(launderrvictim.GetWithBarePtr().Inner)
	})

	// 3. Box pointer in a slice of any, fetch, assert, write
	try("3 []any[i].(*Bare).Field write", func() {
		slot := []any{launderrvictim.GetWithBarePtr().Inner}
		slot[0].(*launderpkg.Bare).Field = "pwnd-via-anyslice"
	})

	// 4. Box, store in a /r/-attacker-local map, fetch, assert, write
	try("4 map[any]→assert→write", func() {
		m := map[string]any{"k": launderrvictim.GetWithBarePtr().Inner}
		m["k"].(*launderpkg.Bare).Field = "pwnd-via-anymap"
	})

	// 5. Read foreign Bare-value field by VALUE (copy), assert it
	//    wasn't somehow useful — should NOT affect victim
	try("5 value copy + write (no effect)", func() {
		b := launderrvictim.GetBareSlice()[0] // value copy
		b.Field = "local-mutate"
		_ = b
	})

	// 6. Box value (not pointer), type-assert, write —
	//    asserting *Bare on a Bare value should fail / not help
	try("6 box-value→assert-ptr", func() {
		var box any = launderrvictim.GetBareSlice()[0] // value, not ptr
		_ = box.(*launderpkg.Bare).Field
	})

	// 7. Pointer aliasing via embedded type's anonymous field
	//    promotion (read the *Bare promoted, write through it)
	try("7 promoted-field ptr write", func() {
		w := launderrvictim.GetWithBareEmbed() // *WithBareEmbed, embedded Bare value
		// w.Field is promoted from embedded Bare
		w.Field = "pwnd-promoted"
	})

	// 8. Promote-via-pointer-embed: WithBareEmbed embeds Bare by
	//    VALUE; take address of promoted Field and write through it
	try("8 &promoted.Field write", func() {
		w := launderrvictim.GetWithBareEmbed()
		p := &w.Field
		*p = "pwnd-promoted-addr"
	})

	println("--- final reads ---")
	println("ptr:", launderrvictim.ReadBarePtr())
	println("embed:", launderrvictim.ReadBareEmbed())
	println("slice[0]:", launderrvictim.ReadBareSlice0())
}

func writeViaAny(x any) {
	t := x.(*launderpkg.Bare)
	t.Field = "pwnd-via-fn-any"
}

// Output:
// 1 any.(*Bare).Field write blocked: true
// 2 fn(any)→assert→write blocked: true
// 3 []any[i].(*Bare).Field write blocked: true
// 4 map[any]→assert→write blocked: true
// 5 value copy + write (no effect) blocked: true
// 6 box-value→assert-ptr blocked: true
// 7 promoted-field ptr write blocked: true
// 8 &promoted.Field write blocked: true
// --- final reads ---
// ptr: bare-ptr-orig
// embed: bare-embed-orig
// slice[0]: slice0

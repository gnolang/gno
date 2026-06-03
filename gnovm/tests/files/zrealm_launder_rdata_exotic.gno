// PKGPATH: gno.land/r/launderrh_exo
package launderrh_exo

// Exotic write patterns against /r/-victim with embedded /p/-Bare.
// No method dispatch on /p/-object, no attacker callbacks.

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
	// 1. Compound assignment on field
	try("1 += on field", func() {
		w := launderrvictim.GetWithBarePtr()
		w.Inner.Field += "-pwnd"
	})

	// 2. Multi-return swap (a, b = b, a) — does each side check readonly?
	try("2 multi-return swap", func() {
		s := launderrvictim.GetBareSlice()
		s[0], s[1] = s[1], s[0]
	})

	// 3. Pointer indirection chain: (*p).Field
	try("3 (*p).Field via star", func() {
		w := launderrvictim.GetWithBarePtr()
		p := w.Inner
		(*p).Field = "pwnd-star"
	})

	// 4. Write inner field of inner struct through full path
	try("4 wp.Inner.Field = ...", func() {
		w := launderrvictim.GetWithBarePtr()
		w.Inner.Field = "pwnd-full-path"
	})

	// 5. Replace whole inner pointer via .Inner =
	try("5 wp.Inner = &fake", func() {
		w := launderrvictim.GetWithBarePtr()
		w.Inner = &launderpkg.Bare{Field: "pwnd-swap"}
	})

	// 6. Slice header overwrite via index-with-capacity slice
	try("6 s[low:high:max] write", func() {
		s := launderrvictim.GetBareSlice()
		s2 := s[0:1:1]
		s2[0].Field = "pwnd-3idx"
	})

	// 7. Assign whole struct: *p = newValue
	try("7 *ptr = newValue", func() {
		w := launderrvictim.GetWithBarePtr()
		*w.Inner = launderpkg.Bare{Field: "pwnd-deref-assign"}
	})

	// 8. Slice-extension append with shared backing
	try("8 append into shared cap", func() {
		s := launderrvictim.GetBareSlice()
		// reslice down then append back — if cap is shared, this
		// overwrites the slot just past len.
		s2 := s[:1]
		s2 = append(s2, launderpkg.Bare{Field: "pwnd-append"})
		_ = s2
	})

	// 9. Read foreign slice/map, store in local, then write through local
	try("9 local alias of foreign", func() {
		alias := launderrvictim.GetBareSlice()
		alias[0].Field = "pwnd-alias"
	})

	// 10. Take address inside foreign struct and pass to a local fn
	try("10 pass addr to local fn", func() {
		w := launderrvictim.GetWithBarePtr()
		writeThrough(&w.Inner.Field)
	})

	println("--- final reads ---")
	println("ptr Inner:", launderrvictim.ReadBarePtr())
	println("slice[0]:", launderrvictim.ReadBareSlice0())
	println("slice[1]:", launderrvictim.ReadBareSlice0Then1())
}

func writeThrough(p *string) {
	*p = "pwnd-via-strptr"
}

// Output:
// 1 += on field blocked: true
// 2 multi-return swap blocked: true
// 3 (*p).Field via star blocked: true
// 4 wp.Inner.Field = ... blocked: true
// 5 wp.Inner = &fake blocked: true
// 6 s[low:high:max] write blocked: true
// 7 *ptr = newValue blocked: true
// 8 append into shared cap blocked: true
// 9 local alias of foreign blocked: true
// 10 pass addr to local fn blocked: true
// --- final reads ---
// ptr Inner: bare-ptr-orig
// slice[0]: slice0
// slice[1]: slice1

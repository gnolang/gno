// PKGPATH: gno.land/r/launderrh_comp
package launderrh_comp

// Composite-type write probes. Victim's /r/-data holds slices,
// arrays, and maps of methods-less /p/-Bare values and pointers.
// Each probe attempts a mutation WITHOUT calling any method on the
// /p/-object and WITHOUT supplying any function value. Pure
// attacker code touching the field.

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
	try("1 slice[i].Field write", func() {
		s := launderrvictim.GetBareSlice()
		s[0].Field = "pwnd-slice-elem-field"
	})
	try("2 slice[i] = bare", func() {
		s := launderrvictim.GetBareSlice()
		s[0] = launderpkg.Bare{Field: "pwnd-slice-swap"}
	})
	try("3 ptrslice[i].Field write", func() {
		s := launderrvictim.GetBarePtrSlice()
		s[0].Field = "pwnd-ptrslice-elem-field"
	})
	try("4 ptrslice[i] = &bare", func() {
		s := launderrvictim.GetBarePtrSlice()
		fake := &launderpkg.Bare{Field: "pwnd-ptr-swap"}
		s[0] = fake
	})
	try("5 arr[i].Field write", func() {
		a := launderrvictim.GetBareArr()
		a[0].Field = "pwnd-arr-elem-field"
	})
	try("6 ptrarr[i].Field write", func() {
		a := launderrvictim.GetBarePtrArr()
		a[0].Field = "pwnd-ptrarr-elem-field"
	})
	try("7 map[k] = bare", func() {
		m := launderrvictim.GetBareMap()
		m["a"] = launderpkg.Bare{Field: "pwnd-map-replace"}
	})
	try("8 ptrmap[k].Field write", func() {
		m := launderrvictim.GetBarePtrMap()
		m["a"].Field = "pwnd-ptrmap-elem-field"
	})
	try("9 ptrmap[k] = &bare", func() {
		m := launderrvictim.GetBarePtrMap()
		fake := &launderpkg.Bare{Field: "pwnd-ptrmap-swap"}
		m["a"] = fake
	})
	try("10 delete(map, k)", func() {
		m := launderrvictim.GetBareMap()
		delete(m, "a")
	})
	try("11 copy(victim, src)", func() {
		s := launderrvictim.GetBareSlice()
		src := []launderpkg.Bare{
			{Field: "pwnd-copy"}, {Field: "pwnd-copy2"},
		}
		copy(s, src)
	})
	try("12 &slice[i].Field write", func() {
		s := launderrvictim.GetBareSlice()
		p := &s[0]
		p.Field = "pwnd-via-elem-addr"
	})
	try("13 subslice[i].Field write", func() {
		s := launderrvictim.GetBareSlice()
		s2 := s[0:1]
		s2[0].Field = "pwnd-via-subslice"
	})

	println("--- final reads ---")
	println("slice[0]:", launderrvictim.ReadBareSlice0())
	println("ptrslice[0]:", launderrvictim.ReadBarePtrSlice0())
	println("arr[0]:", launderrvictim.ReadBareArr0())
	println("ptrarr[0]:", launderrvictim.ReadBarePtrArr0())
	println("map[a]:", launderrvictim.ReadBareMapA())
	println("ptrmap[a]:", launderrvictim.ReadBarePtrMapA())
}

// Output:
// 1 slice[i].Field write blocked: true
// 2 slice[i] = bare blocked: true
// 3 ptrslice[i].Field write blocked: true
// 4 ptrslice[i] = &bare blocked: true
// 5 arr[i].Field write blocked: true
// 6 ptrarr[i].Field write blocked: true
// 7 map[k] = bare blocked: true
// 8 ptrmap[k].Field write blocked: true
// 9 ptrmap[k] = &bare blocked: true
// 10 delete(map, k) blocked: true
// 11 copy(victim, src) blocked: true
// 12 &slice[i].Field write blocked: true
// 13 subslice[i].Field write blocked: true
// --- final reads ---
// slice[0]: slice0
// ptrslice[0]: ptrslice0
// arr[0]: arr0
// ptrarr[0]: ptrarr0
// map[a]: mapA
// ptrmap[a]: ptrmapA

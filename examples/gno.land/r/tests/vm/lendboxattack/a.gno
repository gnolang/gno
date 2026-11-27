package lendboxattack

import (
	lib "gno.land/p/demo/tests/lendboxlib"
	"gno.land/r/tests/vm/lendboxvictim"
)

func KeepBump(cur realm)   { lendboxvictim.AddBox(cross(cur), lib.Bump) }
func KeepSetter(cur realm) { lendboxvictim.AddBox(cross(cur), lib.Setter{V: 777}.Apply) }
func KeepInt(cur realm)    { lendboxvictim.AddCount(cross(cur), lib.BumpInt) }
func KeepOwn(cur realm) {
	lendboxvictim.AddBox(cross(cur), func(b *lib.Box) { b.N += 1000 })
}
func Hook(cur realm)        { lendboxvictim.SetHook(cross(cur), lendboxvictim.Get().Bump) }
func RunLit(cur realm)      { lendboxvictim.RunLit(cross(cur), lendboxvictim.Get().Bump) }
func RunDeferLit(cur realm) { lendboxvictim.RunDeferLit(cross(cur), lendboxvictim.Get().Bump) }

func KeepStored(cur realm)    { lendboxvictim.AddBox(cross(cur), lendboxvictim.DefInit()) }
func KeepStoredInt(cur realm) { lendboxvictim.AddCount(cross(cur), lendboxvictim.DefIntInit()) }

func KeepAdder(cur realm)    { lendboxvictim.AddBox(cross(cur), lendboxvictim.Adder()) }
func KeepIntAdder(cur realm) { lendboxvictim.AddCount(cross(cur), lendboxvictim.IntAdder()) }
func MyAdder(cur realm)      { lendboxvictim.AddBox(cross(cur), lib.MakeAdder(1000)) }

func GuardBump(cur realm)   { lendboxvictim.AddBoxGuarded(cross(cur), lib.Bump) }
func GuardSetter(cur realm) { lendboxvictim.AddBoxGuarded(cross(cur), lib.Setter{V: 777}.Apply) }
func GuardInt(cur realm)    { lendboxvictim.AddCountGuarded(cross(cur), lib.BumpInt) }
func GuardOwn(cur realm) {
	lendboxvictim.AddBoxGuarded(cross(cur), func(b *lib.Box) { b.N += 1000 })
}

func GuardSaved(cur realm) {
	lendboxvictim.AddBoxD(cross(cur), 1)
	lendboxvictim.UpdateGuarded(cross(cur), lib.Bump)
}

package chalbattack

import (
	lib "gno.land/p/demo/tests/chalblib"
	"gno.land/r/tests/vm/chalbvictim"
)

func KeepBump(cur realm)   { chalbvictim.AddBox(cross(cur), lib.Bump) }
func KeepSetter(cur realm) { chalbvictim.AddBox(cross(cur), lib.Setter{V: 777}.Apply) }
func KeepInt(cur realm)    { chalbvictim.AddCount(cross(cur), lib.BumpInt) }
func KeepOwn(cur realm) {
	chalbvictim.AddBox(cross(cur), func(b *lib.Box) { b.N += 1000 })
}
func Hook(cur realm)        { chalbvictim.SetHook(cross(cur), chalbvictim.Get().Bump) }
func RunLit(cur realm)      { chalbvictim.RunLit(cross(cur), chalbvictim.Get().Bump) }
func RunDeferLit(cur realm) { chalbvictim.RunDeferLit(cross(cur), chalbvictim.Get().Bump) }

func KeepStored(cur realm)    { chalbvictim.AddBox(cross(cur), chalbvictim.DefInit()) }
func KeepStoredInt(cur realm) { chalbvictim.AddCount(cross(cur), chalbvictim.DefIntInit()) }

func KeepAdder(cur realm)    { chalbvictim.AddBox(cross(cur), chalbvictim.Adder()) }
func KeepIntAdder(cur realm) { chalbvictim.AddCount(cross(cur), chalbvictim.IntAdder()) }
func MyAdder(cur realm)      { chalbvictim.AddBox(cross(cur), lib.MakeAdder(1000)) }

func GuardBump(cur realm)    { chalbvictim.AddBoxGuarded(cross(cur), lib.Bump) }
func GuardSetter(cur realm)  { chalbvictim.AddBoxGuarded(cross(cur), lib.Setter{V: 777}.Apply) }
func GuardInt(cur realm)     { chalbvictim.AddCountGuarded(cross(cur), lib.BumpInt) }
func GuardOwn(cur realm) {
	chalbvictim.AddBoxGuarded(cross(cur), func(b *lib.Box) { b.N += 1000 })
}

func GuardSaved(cur realm) {
	chalbvictim.AddBoxD(cross(cur), 1)
	chalbvictim.UpdateGuarded(cross(cur), lib.Bump)
}

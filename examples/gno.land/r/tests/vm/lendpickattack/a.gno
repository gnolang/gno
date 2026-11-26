package lendpickattack

import (
	lib "gno.land/p/demo/tests/lendpicklib"
	"gno.land/r/tests/vm/lendpickvictim"
)

func ParamBump(cur realm)  { lendpickvictim.RunParam(cross(cur), lendpickvictim.Get().Bump) }
func ParamLib(cur realm)   { lendpickvictim.RunParamBox(cross(cur), lib.Bump) }
func ParamAdder(cur realm) { lendpickvictim.RunParamBox(cross(cur), lib.MakeAdder(1000)) }
func ParamOwn(cur realm) {
	lendpickvictim.RunParamBox(cross(cur), func(b *lib.Box) { b.N += 1000 })
}
func ResultBump(cur realm) { lendpickvictim.RunResult(cross(cur), lendpickvictim.Get().Bump) }
func PlainBump(cur realm)  { lendpickvictim.RunPlain(cross(cur), lendpickvictim.Get().Bump) }
func RegLib(cur realm)     { lendpickvictim.RunReg(cross(cur), lib.Bump) }
func RegBump(cur realm)    { lendpickvictim.RunRegPlain(cross(cur), lendpickvictim.Get().Bump) }
func IfaceBump(cur realm) {
	lendpickvictim.RunIface(cross(cur), lendpickvictim.Fn(lendpickvictim.Get().Bump))
}
func FieldBump(cur realm) { lendpickvictim.RunField(cross(cur), lendpickvictim.Get().Bump) }

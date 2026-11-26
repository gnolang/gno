package chalcattack

import (
	lib "gno.land/p/demo/tests/chalclib"
	"gno.land/r/tests/vm/chalcvictim"
)

func ParamBump(cur realm)  { chalcvictim.RunParam(cross(cur), chalcvictim.Get().Bump) }
func ParamLib(cur realm)   { chalcvictim.RunParamBox(cross(cur), lib.Bump) }
func ParamAdder(cur realm) { chalcvictim.RunParamBox(cross(cur), lib.MakeAdder(1000)) }
func ParamOwn(cur realm) {
	chalcvictim.RunParamBox(cross(cur), func(b *lib.Box) { b.N += 1000 })
}
func ResultBump(cur realm) { chalcvictim.RunResult(cross(cur), chalcvictim.Get().Bump) }
func PlainBump(cur realm)  { chalcvictim.RunPlain(cross(cur), chalcvictim.Get().Bump) }
func RegLib(cur realm)     { chalcvictim.RunReg(cross(cur), lib.Bump) }
func RegBump(cur realm)    { chalcvictim.RunRegPlain(cross(cur), chalcvictim.Get().Bump) }
func IfaceBump(cur realm)  { chalcvictim.RunIface(cross(cur), chalcvictim.Fn(chalcvictim.Get().Bump)) }
func FieldBump(cur realm)  { chalcvictim.RunField(cross(cur), chalcvictim.Get().Bump) }

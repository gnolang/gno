// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm_b"
)

type Struct struct {
	A int
}

var s *Struct

func init() {
	s2 := &Struct{A: 100}
	withswitch(crossrealm_b.SetObject)(func() { &s2.A })
	s = s2
}

func main() {
	switchrealm()

	s.A = 123
	println(s)
}

// Error:
// cannot modify external-realm or non-realm object

// Preprocessed:
// file{ package crossrealm_test; import crossrealm_b gno.land/r/demo/tests/crossrealm_b; type Struct (const-type gno.land/r/crossrealm_test.Struct); var s<!~VPBlock(2,1)> *(Struct<VPBlock(2,0)>); func init.2() { s2<!~VPBlock(1,0)> := &(Struct<VPBlock(3,0)>{A<VPField(0,0,A)>: (const (100 int))}); (const (withswitch func(func(interface {})) func(interface {})))((const (ref(gno.land/r/demo/tests/crossrealm_b) package{})).SetObject)(func func(){ &(s2<~VPBlock(1,0)>.A) }<s2<()~VPBlock(1,0)>>); s<~VPBlock(3,1)> = s2<~VPBlock(1,0)> }; func main() { (const (switchrealm func()))(); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

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
	cross(crossrealm_b.SetObject)(func() { println(&s2.A) })
	s = s2
}

func main() {
	crossing()

	s.A = 123
	println(s)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): s<~VPBlock(3,1)>.A

// Preprocessed:
// file{ package crossrealm_test; import crossrealm_b gno.land/r/demo/tests/crossrealm_b; type Struct (const-type gno.land/r/crossrealm_test.Struct); var s<!~VPBlock(2,1)> *(typeval{gno.land/r/crossrealm_test.Struct}); func init.2() { s2<!~VPBlock(1,0)> := &(typeval{gno.land/r/crossrealm_test.Struct}{A<VPField(0,0,A)>: (const (100 int))}); (const (cross func(func(interface {})) func(interface {})))((const (ref(gno.land/r/demo/tests/crossrealm_b) package{})).SetObject)(func func(){ (const (println func(...interface {})))(&(s2<~VPBlock(1,0)>.A)) }<s2<()~VPBlock(1,0)>>); s<~VPBlock(3,1)> = s2<~VPBlock(1,0)> }; func main() { (const (crossing func()))(); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

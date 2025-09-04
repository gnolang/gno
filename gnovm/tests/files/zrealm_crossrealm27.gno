// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

var s *Struct

func init() {
	s2 := &Struct{A: 100}
	crossrealm_b.SetObject(cross, func() { println(&s2.A) })
	s = s2
}

func main(cur realm) {
	s.A = 123
	println(s)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): s<~VPBlock(3,1)>.A

// Preprocessed:
// file{ package crossrealm; import crossrealm_b gno.land/r/tests/vm/crossrealm_b; type Struct (const-type gno.land/r/crossrealm.Struct); var s<!~VPBlock(2,1)> *(typeval{gno.land/r/crossrealm.Struct}); func init.2() { s2<!~VPBlock(1,0)> := &((const-type gno.land/r/crossrealm.Struct){A<VPField(0,0,A)>: (const (100 int))}); (const (ref(gno.land/r/tests/vm/crossrealm_b) package{})).SetObject((const (undefined)), func func(){ (const (println func(...interface {})))(&(s2<~VPBlock(1,0)>.A)) }<s2<()~VPBlock(1,0)>>); s<~VPBlock(3,1)> = s2<~VPBlock(1,0)> }; func main(cur (const-type .uverse.realm)) { s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

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
	s = &Struct{A: 100}
	// This is OK, 's' is short for 'crossrelm_test.s',
	// so there are no captures for the closure.
	// Since s is a 'field' of the package block
	// the object 's' is not stored in crossrealm_b,
	// it is lazily attached to this realm!!!
	cross(crossrealm_b.SetObject)(func() { println(&s.A) })
}

func main() {
	crossing()

	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm_test.Struct)

// Preprocessed:
// file{ package crossrealm_test; import crossrealm_b gno.land/r/demo/tests/crossrealm_b; type Struct (const-type gno.land/r/crossrealm_test.Struct); var s<!~VPBlock(2,1)> *(typeval{gno.land/r/crossrealm_test.Struct}); func init.2() { s<~VPBlock(3,1)> = &(typeval{gno.land/r/crossrealm_test.Struct}{A<VPField(0,0,A)>: (const (100 int))}); (const (cross func(func(interface {})) func(interface {})))((const (ref(gno.land/r/demo/tests/crossrealm_b) package{})).SetObject)(func func(){ (const (println func(...interface {})))(&((const (ref(gno.land/r/crossrealm_test) package{})).s.A)) }) }; func main() { (const (crossing func()))(); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

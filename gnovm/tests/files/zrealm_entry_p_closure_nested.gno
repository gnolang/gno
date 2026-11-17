// PKGPATH: gno.land/r/demo/entrynestclosure
package entrynestclosure

import (
	"gno.land/p/demo/tests/nestclosure/v0"
	"gno.land/r/tests/vm/closurehelper"
	"gno.land/r/tests/vm/closurevictim"
)

// A library closure built while its caller ran view-only carries no realm
// stamp; handed to the victim's entry it must not write the victim's data.

func main(cur realm) {
	println("lit:", revive(func() { closurevictim.Run(cross(cur), func(v []int) { v[0] = 1 }) }), closurevictim.Get())
	println("maker:", revive(func() { closurevictim.Run(cross(cur), nestclosure.Maker()) }), closurevictim.Get())
	inner := closurehelper.GetOuter()()
	println("zero-stamp inner:", revive(func() { closurevictim.Run(cross(cur), inner) }), closurevictim.Get())
}

// Output:
// lit: gno.land/r/demo/entrynestclosure's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>[(const (0 int))] 0
// maker: cannot modify object owned by another realm from library code (call a function of the owning realm): v<VPBlock(1,0)>[(const (0 int))] 0
// zero-stamp inner: cannot modify object owned by another realm from library code (call a function of the owning realm): v<VPBlock(1,0)>[(const (0 int))] 0

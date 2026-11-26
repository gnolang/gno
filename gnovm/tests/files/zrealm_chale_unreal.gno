// PKGPATH: gno.land/r/tests/vm/chaleunreal
package chaleunreal

import lib "gno.land/p/demo/tests/chalelib"

var items []*lib.Item
var count int

// A library-owned receiver's method allocates; the realm stores the result.
// Same internal panic with the change reverted: not caused by T55.
func main(cur realm) {
	println("e01", try(func() { items = append(items, lib.Default.Make2(7)) }), len(items))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Error:
// unexpected unreal object: type=*gnolang.HeapItemValue oid=477ce066bfb12f4d4b86fe8a3c34a0accb6463e6:0 isNewReal=false isDirty=false isNewDeleted=false refCount=0

// PKGPATH: gno.land/r/tests/vm/lendregunreal
package lendregunreal

import lib "gno.land/p/demo/tests/lendreglib"

var items []*lib.Item
var count int

// A library-owned receiver's method allocates; the realm stores the result.
// Same internal panic with the change reverted: not caused by the loan rule.
func main(cur realm) {
	println("e01", try(func() { items = append(items, lib.Default.Make2(7)) }), len(items))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Error:
// unexpected unreal object: type=*gnolang.HeapItemValue oid=4c44b07d790fecf0b3a363c718d464aa184aaccb:0 isNewReal=false isDirty=false isNewDeleted=false refCount=0

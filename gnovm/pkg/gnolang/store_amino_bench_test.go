package gnolang

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// storedObjectsForBench returns objects shaped like what an avl-backed realm
// saves on every write: a tree node struct pointing at its children through
// heap items, a heap item holding a ref, and a block of locals.
func storedObjectsForBench() []Object {
	oid := func(n uint64) ObjectID {
		var pid PkgID
		for i := range pid.Hashlet {
			pid.Hashlet[i] = byte(i*7 + 3)
		}
		return ObjectID{PkgID: pid, NewTime: n}
	}
	hash := func(b byte) (h ValueHash) {
		for i := range h.Hashlet {
			h.Hashlet[i] = b + byte(i)
		}
		return
	}
	info := func(id, owner uint64) ObjectInfo {
		return ObjectInfo{ID: oid(id), OwnerID: oid(owner), ModTime: id + 10, RefCount: 1, LastObjectSize: 600}
	}
	node := RefType{ID: "gno.land/p/nt/avl/v0.Node"}
	ref := func(id uint64) TypedValue {
		return TypedValue{T: &PointerType{Elt: node}, V: PointerValue{Base: RefValue{ObjectID: oid(id), Hash: hash(byte(id))}, Index: 0}}
	}
	var height, size TypedValue
	height.T, size.T = Int8Type, IntType
	height.SetInt8(3)
	size.SetInt(42)

	sv := &StructValue{ObjectInfo: info(1000, 999), Fields: []TypedValue{
		{T: StringType, V: StringValue("key-0000012345")},
		{T: StringType, V: StringValue("value-0000012345")},
		height, size,
		ref(1001), ref(1002),
	}}
	hiv := &HeapItemValue{ObjectInfo: info(1001, 1000), Value: TypedValue{T: node, V: RefValue{ObjectID: oid(1003), Hash: hash(9)}}}
	blk := &Block{ObjectInfo: info(1004, 1), Values: []TypedValue{ref(1005), ref(1006), {T: StringType, V: StringValue("k")}, size}}
	return []Object{sv, hiv, blk}
}

// BenchmarkStoredObjectEncode is the encode done by defaultStore.SetObject.
func BenchmarkStoredObjectEncode(b *testing.B) {
	objs := storedObjectsForBench()
	b.ReportAllocs()
	for b.Loop() {
		for _, o := range objs {
			amino.MustMarshalAny(o)
		}
	}
}

// BenchmarkStoredObjectDecode is the decode done by defaultStore.loadObjectSafe.
func BenchmarkStoredObjectDecode(b *testing.B) {
	objs := storedObjectsForBench()
	bzs := make([][]byte, len(objs))
	for i, o := range objs {
		bzs[i] = amino.MustMarshalAny(o)
		var back Object
		amino.MustUnmarshal(bzs[i], &back)
		if !bytes.Equal(amino.MustMarshalAny(back), bzs[i]) {
			b.Fatalf("object %d does not round-trip", i)
		}
	}
	b.ReportAllocs()
	for b.Loop() {
		for _, bz := range bzs {
			var o Object
			amino.MustUnmarshal(bz, &o)
		}
	}
}

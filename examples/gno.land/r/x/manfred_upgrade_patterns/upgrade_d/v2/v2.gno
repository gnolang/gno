package v2

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	v1 "gno.land/r/x/manfred_upgrade_patterns/upgrade_d/v1"
)

var myTree avl.Tree

// MyStruct represents the structure with fields that may have been migrated from v1.
type MyStruct struct {
	FieldA string // Appending "imported-from-v1" if migrating from v1.
	FieldB uint64 // Switching from int to uint64 and multiplying by 1000 if migrating from v1.
	FieldC bool   // Adding a boolean field which is true by default for v1 objects.
}

// String returns a string representation of MyStruct.
func (s *MyStruct) String() string {
	if s == nil {
		return "nil"
	}
	return ufmt.Sprintf("v2:%s:%d:%t", s.FieldA, s.FieldB, s.FieldC)
}

// Get retrieves a MyStruct from the tree by key. If the key does not exist in the current version,
// it attempts to retrieve and migrate the structure from v1.
func Get(key string) *MyStruct {
	ret, ok := myTree.Get(key)
	if !ok {
		v1Struct := v1.Get(key)
		if v1Struct == nil {
			return nil
		}
		// Lazy migration code: convert v1 structure to v2 structure.
		v2Struct := &MyStruct{
			FieldA: v1Struct.FieldA + "-imported-from-v1",
			FieldB: uint64(v1Struct.FieldB * 1000),
			FieldC: true,
		}
		myTree.Set(key, v2Struct)
		return v2Struct
	}
	return ret.(*MyStruct)
}

// init initializes the tree with some default values.
func init() {
	myTree.Set("c", &MyStruct{FieldA: "c", FieldB: 3, FieldC: true})
	myTree.Set("d", &MyStruct{FieldA: "d", FieldB: 4, FieldC: false})
}

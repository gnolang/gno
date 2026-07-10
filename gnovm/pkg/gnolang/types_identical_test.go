package gnolang

import "testing"

func TestIdenticalTypes(t *testing.T) {
	t.Parallel()

	taggedStruct := func(pkgPath string, tag Tag) *StructType {
		return &StructType{
			PkgPath: pkgPath,
			Fields:  []FieldType{{Name: "A", Type: IntType, Tag: tag}},
		}
	}
	method := func(name Name, params ...FieldType) FieldType {
		return FieldType{Name: name, Type: &FuncType{Params: params}}
	}

	tests := []struct {
		name            string
		at, bt          Type
		identical       bool
		identicalNoTags bool
	}{
		{
			name:            "nil types",
			at:              nil,
			bt:              nil,
			identical:       true,
			identicalNoTags: true,
		},
		{
			name:            "nil vs non-nil",
			at:              nil,
			bt:              IntType,
			identical:       false,
			identicalNoTags: false,
		},
		{
			name:            "same primitive",
			at:              IntType,
			bt:              IntType,
			identical:       true,
			identicalNoTags: true,
		},
		{
			name:            "different primitives",
			at:              IntType,
			bt:              Int64Type,
			identical:       false,
			identicalNoTags: false,
		},
		{
			name:            "struct tags differ",
			at:              taggedStruct("main", "a"),
			bt:              taggedStruct("main", "b"),
			identical:       false,
			identicalNoTags: true,
		},
		{
			name:            "struct tags equal",
			at:              taggedStruct("main", "a"),
			bt:              taggedStruct("main", "a"),
			identical:       true,
			identicalNoTags: true,
		},
		{
			name:            "nested struct tags differ through pointer",
			at:              &PointerType{Elt: taggedStruct("main", "a")},
			bt:              &PointerType{Elt: taggedStruct("main", "b")},
			identical:       false,
			identicalNoTags: true,
		},
		{
			name: "embedded vs named field",
			at: &StructType{
				PkgPath: "main",
				Fields:  []FieldType{{Name: "T", Type: IntType, Embedded: true}},
			},
			bt: &StructType{
				PkgPath: "main",
				Fields:  []FieldType{{Name: "T", Type: IntType}},
			},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name: "unexported field from different packages",
			at: &StructType{
				PkgPath: "gno.land/p/demo/a",
				Fields:  []FieldType{{Name: "x", Type: IntType}},
			},
			bt: &StructType{
				PkgPath: "gno.land/p/demo/b",
				Fields:  []FieldType{{Name: "x", Type: IntType}},
			},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name: "exported field from different packages",
			at: &StructType{
				PkgPath: "gno.land/p/demo/a",
				Fields:  []FieldType{{Name: "X", Type: IntType}},
			},
			bt: &StructType{
				PkgPath: "gno.land/p/demo/b",
				Fields:  []FieldType{{Name: "X", Type: IntType}},
			},
			identical:       true,
			identicalNoTags: true,
		},
		{
			name:            "map value types differ",
			at:              &MapType{Key: IntType, Value: IntType},
			bt:              &MapType{Key: IntType, Value: StringType},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name:            "map types equal",
			at:              &MapType{Key: IntType, Value: StringType},
			bt:              &MapType{Key: IntType, Value: StringType},
			identical:       true,
			identicalNoTags: true,
		},
		{
			name: "variadic vs slice param",
			at: &FuncType{Params: []FieldType{
				{Name: "xs", Type: &SliceType{Elt: IntType, Vrd: true}},
			}},
			bt: &FuncType{Params: []FieldType{
				{Name: "xs", Type: &SliceType{Elt: IntType}},
			}},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name: "param names ignored",
			at: &FuncType{Params: []FieldType{
				{Name: "a", Type: &SliceType{Elt: IntType, Vrd: true}},
			}},
			bt: &FuncType{Params: []FieldType{
				{Name: "b", Type: &SliceType{Elt: IntType, Vrd: true}},
			}},
			identical:       true,
			identicalNoTags: true,
		},
		{
			name:            "array lengths differ",
			at:              &ArrayType{Len: 2, Elt: IntType},
			bt:              &ArrayType{Len: 3, Elt: IntType},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name:            "slice types equal",
			at:              &SliceType{Elt: IntType},
			bt:              &SliceType{Elt: IntType},
			identical:       true,
			identicalNoTags: true,
		},
		{
			name: "interface method order ignored",
			at: &InterfaceType{
				PkgPath: "main",
				Methods: []FieldType{method("A"), method("B")},
			},
			bt: &InterfaceType{
				PkgPath: "main",
				Methods: []FieldType{method("B"), method("A")},
			},
			identical:       true,
			identicalNoTags: true,
		},
		{
			name: "interface method signatures differ",
			at: &InterfaceType{
				PkgPath: "main",
				Methods: []FieldType{
					method("M", FieldType{Type: &SliceType{Elt: IntType, Vrd: true}}),
				},
			},
			bt: &InterfaceType{
				PkgPath: "main",
				Methods: []FieldType{
					method("M", FieldType{Type: &SliceType{Elt: IntType}}),
				},
			},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name: "unexported method from different packages",
			at: &InterfaceType{
				PkgPath: "gno.land/p/demo/a",
				Methods: []FieldType{method("m")},
			},
			bt: &InterfaceType{
				PkgPath: "gno.land/p/demo/b",
				Methods: []FieldType{method("m")},
			},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name:            "chan directions differ",
			at:              &ChanType{Dir: SEND, Elt: IntType},
			bt:              &ChanType{Dir: RECV, Elt: IntType},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name:            "chan types equal",
			at:              &ChanType{Dir: SEND | RECV, Elt: IntType},
			bt:              &ChanType{Dir: SEND | RECV, Elt: IntType},
			identical:       true,
			identicalNoTags: true,
		},
		{
			name:            "tuple types differ",
			at:              &tupleType{Elts: []Type{IntType, StringType}},
			bt:              &tupleType{Elts: []Type{IntType, IntType}},
			identical:       false,
			identicalNoTags: false,
		},
		{
			name:            "different kinds",
			at:              &SliceType{Elt: IntType},
			bt:              &ArrayType{Len: 1, Elt: IntType},
			identical:       false,
			identicalNoTags: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := identicalTypes(tc.at, tc.bt); got != tc.identical {
				t.Errorf("identicalTypes(%v, %v) = %v, want %v",
					tc.at, tc.bt, got, tc.identical)
			}
			// identity must be symmetric.
			if got := identicalTypes(tc.bt, tc.at); got != tc.identical {
				t.Errorf("identicalTypes(%v, %v) = %v, want %v (symmetry)",
					tc.bt, tc.at, got, tc.identical)
			}
			if got := identicalTypesIgnoreTags(tc.at, tc.bt); got != tc.identicalNoTags {
				t.Errorf("identicalTypesIgnoreTags(%v, %v) = %v, want %v",
					tc.at, tc.bt, got, tc.identicalNoTags)
			}
			if got := identicalTypesIgnoreTags(tc.bt, tc.at); got != tc.identicalNoTags {
				t.Errorf("identicalTypesIgnoreTags(%v, %v) = %v, want %v (symmetry)",
					tc.bt, tc.at, got, tc.identicalNoTags)
			}
		})
	}
}

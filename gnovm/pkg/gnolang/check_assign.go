package gnolang

// Assignability rules the preprocessor enforces (Go spec: "Type
// identity", "Assignment statements", "Address operators").

// tagsIdentical reports whether two types with the same TypeID also
// agree on struct tags, which TypeID leaves out but type identity
// includes. Declared types are identical by name alone.
func tagsIdentical(a, b Type) bool {
	switch at := a.(type) {
	case *StructType:
		bt, ok := b.(*StructType)
		if !ok || len(at.Fields) != len(bt.Fields) {
			return false
		}
		for i := range at.Fields {
			if at.Fields[i].Tag != bt.Fields[i].Tag ||
				!tagsIdentical(at.Fields[i].Type, bt.Fields[i].Type) {
				return false
			}
		}
	case *PointerType:
		if bt, ok := b.(*PointerType); ok {
			return tagsIdentical(at.Elt, bt.Elt)
		}
	case *ArrayType:
		if bt, ok := b.(*ArrayType); ok {
			return tagsIdentical(at.Elt, bt.Elt)
		}
	case *SliceType:
		if bt, ok := b.(*SliceType); ok {
			return tagsIdentical(at.Elt, bt.Elt)
		}
	case *MapType:
		if bt, ok := b.(*MapType); ok {
			return tagsIdentical(at.Key, bt.Key) && tagsIdentical(at.Value, bt.Value)
		}
	case *FuncType:
		if bt, ok := b.(*FuncType); ok {
			return fieldTagsIdentical(at.Params, bt.Params) && fieldTagsIdentical(at.Results, bt.Results)
		}
	}
	return true
}

func fieldTagsIdentical(a, b []FieldType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !tagsIdentical(a[i].Type, b[i].Type) {
			return false
		}
	}
	return true
}

// inMapValue reports whether lx is a struct field or array element of a
// value held in a map, reached without a pointer: it is not addressable.
func inMapValue(store Store, last BlockNode, lx Expr) bool {
	var inner Expr
	switch x := lx.(type) {
	case *SelectorExpr:
		if _, ok := baseOf(evalStaticTypeOf(store, last, x.X)).(*StructType); !ok {
			return false
		}
		inner = x.X
	case *IndexExpr:
		if _, ok := baseOf(evalStaticTypeOf(store, last, x.X)).(*ArrayType); !ok {
			return false
		}
		inner = x.X
	default:
		return false
	}
	if ix, ok := inner.(*IndexExpr); ok {
		if _, ok := baseOf(evalStaticTypeOf(store, last, ix.X)).(*MapType); ok {
			return true
		}
	}
	return inMapValue(store, last, inner)
}

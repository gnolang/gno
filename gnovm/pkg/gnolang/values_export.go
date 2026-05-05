package gnolang

import (
	"fmt"
	"reflect"
	"strconv"
)

// ExportRefValue represents a back-reference to an ephemeral Object already
// emitted earlier in the export stream. Unlike RefValue (which uses an
// ObjectID for persisted objects), ExportRefValue uses a synthetic ":N" ID,
// where N is an incrementing counter assigned in the encoder's DFS traversal
// order. The first time an ephemeral Object is visited it is expanded inline
// and assigned N = (count of previously-seen ephemeral Objects) + 1; any
// subsequent visit to the same Object emits ExportRefValue{":N"} instead of
// re-expanding it.
//
// Consumers that need to resolve ":N" back to its inline occurrence must walk
// the exported tree in the same DFS order the encoder uses (source-order
// fields, slice/array indices, MapList insertion order, Block values), count
// each inline ephemeral Object as they encounter it, and look up the Nth one.
// Registered with Amino as "/gno.ExportRefValue".
type ExportRefValue struct {
	ObjectID string `json:"ObjectID"` // ":1", ":2", etc.
}

func (ExportRefValue) assertValue() {}
func (erv ExportRefValue) String() string {
	return fmt.Sprintf("exportref(%s)", erv.ObjectID)
}
func (erv ExportRefValue) VisitAssociated(_ Visitor) (stop bool) { return false }

// GetShallowSize returns the size of an ExportRefValue for alloc-gas
// accounting. Uses the same size as RefValue since they're structurally
// equivalent (one ObjectID string).
func (erv ExportRefValue) GetShallowSize() int64 { return allocRefValue }

// ExportValues exports multiple TypedValues for JSON serialization.
// It walks the value tree and:
//   - Replaces persisted (real) objects with RefValue{ObjectID: ...}
//   - Breaks cycles in ephemeral (unreal) objects with ExportRefValue{ObjectID: ":N"}
//   - Copies all values defensively to prevent accidental mutation
//
// The result is suitable for amino.MarshalJSON() serialization.
func ExportValues(tvs []TypedValue) []TypedValue {
	seen := make(map[Object]int)
	result := make([]TypedValue, len(tvs))
	for i, tv := range tvs {
		result[i] = exportValue(tv, seen)
	}
	return result
}

// ExportObject exports a single Object for JSON serialization.
// The object is expanded inline (depth 0), but nested real objects
// become RefValue references. Ephemeral cycles are broken with
// ExportRefValue{":N"} references.
func ExportObject(obj Object) Value {
	seen := make(map[Object]int)
	return exportObjectToValue(obj, seen)
}

// exportValue exports a TypedValue, replacing objects with refs.
func exportValue(tv TypedValue, seen map[Object]int) TypedValue {
	result := TypedValue{N: tv.N}
	if tv.T != nil {
		result.T = exportRefOrCopyType(tv.T, seen)
	}
	if obj, ok := tv.V.(Object); ok {
		result.V = exportToRefOrCopy(obj, seen)
		return result
	}
	if tv.V != nil {
		result.V = exportCopyValue(tv.V, seen)
	}
	return result
}

// exportObjectToValue exports an Object, expanding it inline.
// Nested real objects become RefValue. Ephemeral cycles are broken.
func exportObjectToValue(obj Object, seen map[Object]int) Value {
	if obj == nil {
		return nil
	}

	// Unwrap HeapItemValue: if the inner value is an Object (ephemeral case),
	// export the inner object directly. For persisted HeapItemValues, the inner
	// value is a RefValue (not an Object), so this is a no-op.
	if hiv, ok := obj.(*HeapItemValue); ok {
		if innerObj, ok := hiv.Value.V.(Object); ok {
			obj = innerObj
		}
	}

	// Check for cycles
	if id, exists := seen[obj]; exists {
		if obj.GetIsReal() {
			return RefValue{
				ObjectID: obj.GetObjectID(),
			}
		}
		return ExportRefValue{
			ObjectID: ":" + strconv.Itoa(id),
		}
	}

	// Mark seen
	id := len(seen) + 1
	seen[obj] = id

	// Expand inline
	return exportCopyValue(obj, seen)
}

// exportToRefOrCopy converts an Object to a RefValue if it's persisted,
// or copies it inline if it's ephemeral.
// This is analogous to realm.go's toRefValue but handles unreal objects
// by assigning synthetic cycle-breaking IDs instead of panicking.
func exportToRefOrCopy(val Value, seen map[Object]int) Value {
	if ref, ok := val.(RefValue); ok {
		return ref
	}

	oo, ok := val.(Object)
	if !ok {
		panic("unexpected error converting to ref value")
	}

	// Packages always become refs
	if pv, ok := val.(*PackageValue); ok {
		return RefValue{PkgPath: pv.PkgPath}
	}

	// Real (persisted) objects always become RefValue with their real ObjectID.
	// Their children are already RefValues in the store, so cycles are impossible.
	if oo.GetIsReal() {
		return RefValue{
			ObjectID: oo.GetObjectID(),
			Hash:     oo.GetHash(),
		}
	}

	// Unreal (ephemeral) objects: check for cycles
	if id, exists := seen[oo]; exists {
		return ExportRefValue{
			ObjectID: ":" + strconv.Itoa(id),
		}
	}

	// Not yet seen: assign ID, copy inline
	id := len(seen) + 1
	seen[oo] = id
	return exportCopyValue(oo, seen)
}

// exportCopyValue creates a defensive copy of a Value with refs for objects.
// This mirrors realm.go's copyValueWithRefs but handles unreal objects.
func exportCopyValue(val Value, seen map[Object]int) Value {
	switch cv := val.(type) {
	case nil:
		return nil
	case StringValue:
		return cv
	case BigintValue:
		return cv
	case BigdecValue:
		return cv
	case DataByteValue:
		panic("cannot copy data byte value")
	case PointerValue:
		if cv.Base == nil {
			panic("pointer with nil base")
		}
		return PointerValue{
			Base:  exportToRefOrCopy(cv.Base, seen),
			Index: cv.Index,
		}
	case *ArrayValue:
		if cv.Data == nil {
			list := make([]TypedValue, len(cv.List))
			for i, etv := range cv.List {
				list[i] = exportValue(etv, seen)
			}
			return &ArrayValue{
				ObjectInfo: cv.ObjectInfo.Copy(),
				List:       list,
			}
		}
		return &ArrayValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Data:       cp(cv.Data),
		}
	case *SliceValue:
		return &SliceValue{
			Base:   exportToRefOrCopy(cv.Base, seen),
			Offset: cv.Offset,
			Length: cv.Length,
			Maxcap: cv.Maxcap,
		}
	case *StructValue:
		fields := make([]TypedValue, len(cv.Fields))
		for i, ftv := range cv.Fields {
			fields[i] = exportValue(ftv, seen)
		}
		return &StructValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Fields:     fields,
		}
	case *FuncValue:
		source := toRefNode(cv.Source)
		var parent Value
		if cv.Parent != nil {
			parent = exportToRefOrCopy(cv.Parent, seen)
		}
		captures := make([]TypedValue, len(cv.Captures))
		for i, ctv := range cv.Captures {
			captures[i] = exportValue(ctv, seen)
		}
		if cv.nativeBody != nil && cv.NativePkg == "" {
			panic("cannot copy function value with native body when there is no native package")
		}
		ft := exportCopyTypeWithRefs(cv.Type, seen)
		return &FuncValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Type:       ft,
			IsMethod:   cv.IsMethod,
			Source:     source,
			Name:       cv.Name,
			Parent:     parent,
			Captures:   captures,
			FileName:   cv.FileName,
			PkgPath:    cv.PkgPath,
			NativePkg:  cv.NativePkg,
			NativeName: cv.NativeName,
			Crossing:   cv.Crossing,
		}
	case *BoundMethodValue:
		// cv.Func is typed *FuncValue, not Value, so it can't carry a
		// RefValue/ExportRefValue back-reference. This mirrors realm.go's
		// copyValueWithRefs pattern. Safe because a BoundMethodValue holds
		// a unique, freshly-constructed FuncValue instance that is not
		// shared with any other traversal path (BoundMethodValue is
		// created at bind time, not deduplicated). If that invariant ever
		// changes, this branch would re-expand a shared FuncValue inline
		// instead of emitting an ExportRefValue — the exported output
		// would still be correct, just potentially larger.
		fnc := exportCopyValue(cv.Func, seen).(*FuncValue)
		rtv := exportValue(cv.Receiver, seen)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			key2 := exportValue(cur.Key, seen)
			val2 := exportValue(cur.Value, seen)
			list.Append(nilAllocator, key2).Value = val2
		}
		return &MapValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			List:       list,
		}
	case TypeValue:
		// Export the type as a reference, not inline. Consumers that
		// need the full definition (e.g. struct field names, method set)
		// resolve the TypeID via the vm/qtype_json query endpoint.
		// Keeping this symmetric with field-position types (which also
		// go through exportRefOrCopyType at Layer 1) gives a uniform
		// wire shape and smaller JSON payloads.
		return toTypeValue(exportRefOrCopyType(cv.Type, seen))
	case *PackageValue:
		return RefValue{PkgPath: cv.PkgPath}
	case *Block:
		source := toRefNode(cv.Source)
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = exportValue(tv, seen)
		}
		var bparent Value
		if cv.Parent != nil {
			bparent = exportToRefOrCopy(cv.Parent, seen)
		}
		return &Block{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Source:     source,
			Values:     vals,
			Parent:     bparent,
			Blank:      TypedValue{},
		}
	case RefValue:
		return cv
	case *HeapItemValue:
		return &HeapItemValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Value:      exportValue(cv.Value, seen),
		}
	default:
		panic(fmt.Sprintf("unexpected type %v", reflect.TypeOf(val)))
	}
}

// exportRefOrCopyType replaces DeclaredTypes with RefType, copies others.
func exportRefOrCopyType(typ Type, seen map[Object]int) Type {
	if dt, ok := typ.(*DeclaredType); ok {
		return RefType{ID: dt.TypeID()}
	}
	return exportCopyTypeWithRefs(typ, seen)
}

// exportCopyTypeWithRefs copies a type, replacing DeclaredTypes with RefType.
func exportCopyTypeWithRefs(typ Type, seen map[Object]int) Type {
	switch ct := typ.(type) {
	case nil:
		panic("cannot copy nil types")
	case PrimitiveType:
		return ct
	case *PointerType:
		return &PointerType{
			Elt: exportRefOrCopyType(ct.Elt, seen),
		}
	case FieldType:
		panic("cannot copy field types")
	case *ArrayType:
		return &ArrayType{
			Len: ct.Len,
			Elt: exportRefOrCopyType(ct.Elt, seen),
			Vrd: ct.Vrd,
		}
	case *SliceType:
		return &SliceType{
			Elt: exportRefOrCopyType(ct.Elt, seen),
			Vrd: ct.Vrd,
		}
	case *StructType:
		return &StructType{
			PkgPath: ct.PkgPath,
			Fields:  exportCopyFieldsWithRefs(ct.Fields, seen),
		}
	case *FuncType:
		return &FuncType{
			Params:  exportCopyFieldsWithRefs(ct.Params, seen),
			Results: exportCopyFieldsWithRefs(ct.Results, seen),
		}
	case *MapType:
		return &MapType{
			Key:   exportRefOrCopyType(ct.Key, seen),
			Value: exportRefOrCopyType(ct.Value, seen),
		}
	case *InterfaceType:
		return &InterfaceType{
			PkgPath: ct.PkgPath,
			Methods: exportCopyFieldsWithRefs(ct.Methods, seen),
			Generic: ct.Generic,
		}
	case *TypeType:
		return &TypeType{}
	case *DeclaredType:
		// Likely dead code. Every path that could hand a *DeclaredType
		// to this function now routes through exportRefOrCopyType at
		// Layer 1 instead, which collapses DeclaredTypes to RefType{ID}
		// before reaching this switch: field/element types via
		// exportCopyFieldsWithRefs, TypeValue positions in
		// exportCopyValue, and tv.T in exportValue. *FuncValue.Type and
		// method mtv.T are *FuncType, never DeclaredType.
		// DeclaredType.Base is invariantly non-DeclaredType per the
		// types.go:1441 doc comment (enforced by declareWith/baseOf).
		// Kept as defensive code; if a future caller hands in a
		// DeclaredType directly, the inlined form here + exportCopyMethods
		// below both still produce correct output.
		dt := &DeclaredType{
			PkgPath: ct.PkgPath,
			Name:    ct.Name,
			Base:    exportCopyTypeWithRefs(ct.Base, seen),
			Methods: exportCopyMethods(ct.Methods, seen),
		}
		return dt
	case *PackageType:
		return &PackageType{}
	case *ChanType:
		return &ChanType{
			Dir: ct.Dir,
			Elt: exportRefOrCopyType(ct.Elt, seen),
		}
	case blockType:
		return blockType{}
	case *tupleType:
		elts2 := make([]Type, len(ct.Elts))
		for i, elt := range ct.Elts {
			elts2[i] = exportRefOrCopyType(elt, seen)
		}
		return &tupleType{
			Elts: elts2,
		}
	case RefType:
		return RefType{ID: ct.ID}
	case heapItemType:
		return ct
	default:
		panic(fmt.Sprintf("unexpected type %v", typ))
	}
}

func exportCopyFieldsWithRefs(fields []FieldType, seen map[Object]int) []FieldType {
	fieldsCpy := make([]FieldType, len(fields))
	for i, field := range fields {
		fieldsCpy[i] = FieldType{
			Name:     field.Name,
			Type:     exportRefOrCopyType(field.Type, seen),
			Embedded: field.Embedded,
			Tag:      field.Tag,
		}
	}
	return fieldsCpy
}

// exportCopyMethods is reached only from the *DeclaredType branch of
// exportCopyTypeWithRefs, which is itself likely dead code post-fix
// (see comment there). Kept as defensive code. One caveat if it does
// ever fire: V is expanded via exportCopyValue rather than
// exportToRefOrCopy, so if the same *FuncValue is reachable elsewhere
// in the exported tree (e.g. via a BoundMethodValue holding it), it
// gets re-expanded inline rather than deduplicated. Acceptable because
// the inlined copies are byte-identical; a consumer sees duplication
// but not inconsistency.
func exportCopyMethods(methods []TypedValue, seen map[Object]int) []TypedValue {
	res := make([]TypedValue, len(methods))
	for i, mtv := range methods {
		res[i] = TypedValue{
			T: exportCopyTypeWithRefs(mtv.T, seen),
			V: exportCopyValue(mtv.V, seen),
		}
	}
	return res
}

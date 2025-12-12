package gnolang

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// MUST NOT modify anything inside tv.
func ExportValues(tvs []TypedValue) []TypedValue {
	seens := map[Object]int{}
	ntvs := make([]TypedValue, len(tvs))
	for i, tv := range tvs {
		ntvs[i] = exportValue(tv, seens)
	}

	return ntvs
}

// MUST NOT modify anything inside tv.
func ExportValue(tv TypedValue) TypedValue {
	return exportValue(tv, map[Object]int{})
}

func exportValue(tv TypedValue, seen map[Object]int) TypedValue {
	if tv.T != nil {
		tv.T = exportRefOrCopyType(tv.T, seen)
	}

	if obj, ok := tv.V.(Object); ok {
		tv.V = exportToValueOrRefValue(obj, seen)
		return tv
	}

	tv.V = exportCopyValueWithRefs(tv.V, seen)
	return tv
}

// Copies value but with references to objects; the result is suitable for
// persistence bytes serialization.
// Also checks for integrity of immediate children -- they must already be
// persistent (real), and not dirty, or else this function panics.
func exportCopyValueWithRefs(val Value, m map[Object]int) Value {
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
		panic("cannot copy data byte value with references")
	case PointerValue:
		if cv.Base == nil {
			panic("should not happen")
		}
		return PointerValue{
			/*
				already represented in .Base[Index]:
				TypedValue: &TypedValue{
					T: cv.TypedValue.T,
					V: copyValueWithRefs(cv.TypedValue.V),
				},
			*/
			Base:  exportToValueOrRefValue(cv.Base, m),
			Index: cv.Index,
		}
	case *ArrayValue:
		if cv.Data == nil {
			list := make([]TypedValue, len(cv.List))
			for i, etv := range cv.List {
				list[i] = exportValue(etv, m)
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
			Base:   exportToValueOrRefValue(cv.Base, m),
			Offset: cv.Offset,
			Length: cv.Length,
			Maxcap: cv.Maxcap,
		}
	case *StructValue:
		fields := make([]TypedValue, len(cv.Fields))
		for i, ftv := range cv.Fields {
			fields[i] = exportValue(ftv, m)
		}
		return &StructValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Fields:     fields,
		}
	case *FuncValue:
		source := toRefNode(cv.Source)
		var parent Value
		if cv.Parent != nil {
			parent = exportToValueOrRefValue(cv.Parent, m)
		}
		captures := make([]TypedValue, len(cv.Captures))
		for i, ctv := range cv.Captures {
			captures[i] = exportValue(ctv, m)
		}
		// nativeBody funcs which don't come from NativeResolver (and thus don't
		// have NativePkg/Name) can't be persisted, and should not be able
		// to get here anyway.
		if cv.nativeBody != nil && cv.NativePkg == "" {
			panic("cannot copy function value with native body when there is no native package")
		}
		ft := exportCopyTypeWithRefs(cv.Type, m)
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
		fnc := exportCopyValueWithRefs(cv.Func, m).(*FuncValue)
		rtv := exportValue(cv.Receiver, m)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			key2 := exportValue(cur.Key, m)
			val2 := exportValue(cur.Value, m)
			list.Append(nilAllocator, key2).Value = val2
		}
		return &MapValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			List:       list,
		}
	case TypeValue:
		return toTypeValue(exportCopyTypeWithRefs(cv.Type, m))
	case *PackageValue:
		block := exportToValueOrRefValue(cv.Block, m)
		fblocks := make([]Value, len(cv.FBlocks))
		for i, fb := range cv.FBlocks {
			fblocks[i] = exportToValueOrRefValue(fb, m)
		}
		return &PackageValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Block:      block,
			PkgName:    cv.PkgName,
			PkgPath:    cv.PkgPath,
			FNames:     cv.FNames, // no copy
			FBlocks:    fblocks,
			Realm:      cv.Realm,
		}
	case *Block:
		source := toRefNode(cv.Source)
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = exportValue(tv, m)
		}
		var bparent Value
		if cv.Parent != nil {
			bparent = exportToValueOrRefValue(cv.Parent, m)
		}
		bl := &Block{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Source:     source,
			Values:     vals,
			Parent:     bparent,
			Blank:      TypedValue{}, // empty
		}
		return bl
	case RefValue:
		return cv
	case *HeapItemValue:
		hiv := &HeapItemValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Value:      exportValue(cv.Value, m),
		}
		return hiv
	default:
		panic(fmt.Sprintf(
			"unexpected type %v", reflect.TypeOf(val),
		))
	}
}

type seenObject map[Object]int

func (s seenObject) has(o Object) (ok bool) {
	_, ok = s[o]
	return ok
}

func (s seenObject) add(o Object) (id int, new bool) {
	var exist bool

	// Create a local reference to Object for export
	id, exist = s[o]
	if new = !exist; new {
		id = len(s) + 1 // Avoid empty 0 value
		s[o] = id
	}

	return
}

func exportToValueOrRefValue(val Value, seen seenObject) Value {
	// TODO use type switch stmt.
	if ref, ok := val.(RefValue); ok {
		return ref
	}

	oo, ok := val.(Object)
	if !ok {
		panic("unexpected error converting to ref value")
	}

	if pv, ok := val.(*PackageValue); ok {
		if pv.GetIsDirty() {
			panic("unexpected dirty package " + pv.PkgPath)
		}
		return RefValue{
			PkgPath: pv.PkgPath,
		}
	}

	if !oo.GetIsReal() {
		if id, ok := seen[oo]; ok {
			return RefValue{
				ObjectID: ObjectID{NewTime: uint64(id)},
				Escaped:  true,
			}
		}

		id := len(seen) + 1
		seen[oo] = id
		no := exportCopyValueWithRefs(oo, seen).(Object)
		no.SetObjectID(ObjectID{NewTime: uint64(id)})
		seen[no] = id

		return RefValue{
			ObjectID: no.GetObjectID(),
			Escaped:  true,
		}
	}

	if id, ok := seen[oo]; ok { // Check if we already have seen this object locally
		return RefValue{
			ObjectID: ObjectID{NewTime: uint64(id)},
			Escaped:  true,
		}
	}

	if oo.GetIsDirty() {
		// This can happen with some circular
		// references.
		// panic("unexpected dirty object")
	}

	if oo.GetIsNewEscaped() {
		// NOTE: oo.GetOwnerID() will become zero.
		return RefValue{
			ObjectID: oo.GetObjectID(),
			Escaped:  true,
			// Hash: nil,
		}
	}

	if oo.GetIsEscaped() {
		if debug {
			if !oo.GetOwnerID().IsZero() {
				panic("cannot convert escaped object to ref value without an owner ID")
			}
		}
		return RefValue{
			ObjectID: oo.GetObjectID(),
			Escaped:  true,
			// Hash: nil,
		}
	}

	if debug {
		if oo.GetRefCount() > 1 {
			panic("unexpected references when converting to ref value")
		}
		if oo.GetHash().IsZero() {
			panic("hash missing when converting to ref value")
		}
	}
	return RefValue{
		ObjectID: oo.GetObjectID(),
		Hash:     oo.GetHash(),
	}
}

func exportRefOrCopyType(typ Type, m map[Object]int) Type {
	if dt, ok := typ.(*DeclaredType); ok {
		return RefType{ID: dt.TypeID()}
	} else {
		return exportCopyTypeWithRefs(typ, m)
	}
}

// the result is suitable for persistence bytes serialization.
func exportCopyTypeWithRefs(typ Type, m map[Object]int) Type {
	switch ct := typ.(type) {
	case nil:
		panic("cannot copy nil types")
	case PrimitiveType:
		return ct
	case *PointerType:
		return &PointerType{
			Elt: exportRefOrCopyType(ct.Elt, m),
		}
	case FieldType:
		panic("cannot copy field types")
	case *ArrayType:
		return &ArrayType{
			Len: ct.Len,
			Elt: exportRefOrCopyType(ct.Elt, m),
			Vrd: ct.Vrd,
		}
	case *SliceType:
		return &SliceType{
			Elt: exportRefOrCopyType(ct.Elt, m),
			Vrd: ct.Vrd,
		}
	case *StructType:
		return &StructType{
			PkgPath: ct.PkgPath,
			Fields:  exportCopyFieldsWithRefs(ct.Fields, m),
		}
	case *FuncType:
		return &FuncType{
			Params:  exportCopyFieldsWithRefs(ct.Params, m),
			Results: exportCopyFieldsWithRefs(ct.Results, m),
		}
	case *MapType:
		return &MapType{
			Key:   exportRefOrCopyType(ct.Key, m),
			Value: exportRefOrCopyType(ct.Value, m),
		}
	case *InterfaceType:
		return &InterfaceType{
			PkgPath: ct.PkgPath,
			Methods: exportCopyFieldsWithRefs(ct.Methods, m),
			Generic: ct.Generic,
		}
	case *TypeType:
		return &TypeType{}
	case *DeclaredType:
		dt := &DeclaredType{
			PkgPath: ct.PkgPath,
			Name:    ct.Name,
			Base:    exportCopyTypeWithRefs(ct.Base, m),
			Methods: exportCopyMethods(ct.Methods, m),
		}
		return dt
	case *PackageType:
		return &PackageType{}
	case *ChanType:
		return &ChanType{
			Dir: ct.Dir,
			Elt: exportRefOrCopyType(ct.Elt, m),
		}
	case blockType:
		return blockType{}
	case *tupleType:
		elts2 := make([]Type, len(ct.Elts))
		for i, elt := range ct.Elts {
			elts2[i] = exportRefOrCopyType(elt, m)
		}
		return &tupleType{
			Elts: elts2,
		}
	case RefType:
		return RefType{
			ID: ct.ID,
		}
	case heapItemType:
		return ct
	default:
		panic(fmt.Sprintf(
			"unexpected type %v", typ))
	}
}

func exportCopyFieldsWithRefs(fields []FieldType, m map[Object]int) []FieldType {
	fieldsCpy := make([]FieldType, len(fields))
	for i, field := range fields {
		fieldsCpy[i] = FieldType{
			Name:     field.Name,
			Type:     exportRefOrCopyType(field.Type, m),
			Embedded: field.Embedded,
			Tag:      field.Tag,
		}
	}
	return fieldsCpy
}

func exportCopyMethods(methods []TypedValue, m map[Object]int) []TypedValue {
	res := make([]TypedValue, len(methods))
	for i, mtv := range methods {
		// NOTE: this works because copyMethods/copyTypeWithRefs
		// gets called AFTER the file block (method closure)
		// gets saved (e.g. from *Machine.savePackage()).
		res[i] = TypedValue{
			T: exportCopyTypeWithRefs(mtv.T, m),
			V: exportCopyValueWithRefs(mtv.V, m),
		}
	}
	return res
}

type jsonTypedValue struct {
	Type  json.RawMessage `json:"T"`
	Value json.RawMessage `json:"V"`
}

func JSONExportTypedValues(tvs []TypedValue, seen map[Object]int) ([]byte, error) {
	if seen == nil {
		seen = map[Object]int{}
	}

	jexps := make([]*jsonTypedValue, len(tvs))

	for i, tv := range tvs {
		jexps[i] = jsonExportedTypedValue(exportValue(tv, seen))
	}

	return json.Marshal(jexps)
}

func JSONExportTypedValue(tv TypedValue, seen map[Object]int) ([]byte, error) {
	if seen == nil {
		seen = map[Object]int{}
	}

	tv = exportValue(tv, seen) // first export value
	return json.Marshal(jsonExportedTypedValue(tv))
}

func jsonExportedTypedValue(tv TypedValue) (exp *jsonTypedValue) {
	return &jsonTypedValue{
		Type:  jsonExportedType(tv.T),
		Value: jsonExportedValue(tv),
	}
}

func jsonExportedType(typ Type) []byte {
	if typ == nil {
		return []byte("null")
	}

	var ret string
	switch ct := typ.(type) {
	case RefType:
		ret = ct.TypeID().String()
	default:
		ret = ct.String()
	}

	return []byte(strconv.Quote(ret))
}

func jsonExportedValue(tv TypedValue) []byte {
	bt := BaseOf(tv.T)
	switch bt := bt.(type) {
	case PrimitiveType:
		var ret string
		switch bt {
		case UntypedBoolType, BoolType:
			ret = strconv.FormatBool(tv.GetBool())
		case UntypedStringType, StringType:
			ret = strconv.Quote(tv.GetString())
		case IntType:
			ret = fmt.Sprintf("%d", tv.GetInt())
		case Int8Type:
			ret = fmt.Sprintf("%d", tv.GetInt8())
		case Int16Type:
			ret = fmt.Sprintf("%d", tv.GetInt16())
		case UntypedRuneType, Int32Type:
			ret = fmt.Sprintf("%d", tv.GetInt32())
		case Int64Type:
			ret = fmt.Sprintf("%d", tv.GetInt64())
		case UintType:
			ret = fmt.Sprintf("%d", tv.GetUint())
		case Uint8Type:
			ret = fmt.Sprintf("%d", tv.GetUint8())
		case DataByteType:
			ret = fmt.Sprintf("%d", tv.GetDataByte())
		case Uint16Type:
			ret = fmt.Sprintf("%d", tv.GetUint16())
		case Uint32Type:
			ret = fmt.Sprintf("%d", tv.GetUint32())
		case Uint64Type:
			ret = fmt.Sprintf("%d", tv.GetUint64())
		case Float32Type:
			ret = fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32()))
		case Float64Type:
			ret = fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64()))
		case UntypedBigintType:
			ret = tv.V.(BigintValue).V.String()
		case UntypedBigdecType:
			ret = tv.V.(BigdecValue).V.String()
		default:
			panic("invalid primitive type - should not happen")
		}

		return []byte(ret)
	default:
		if tv.V == nil {
			return []byte("null")
		}

		return amino.MustMarshalJSONAny(tv.V)
	}
}

// Constants for depth limiting in JSON export
const (
	DefaultMaxDepth = 3  // Default depth limit for JSON export
	UnlimitedDepth  = -1 // No depth limit (use with caution)
)

// JSONTypedValue represents a human-readable JSON format for TypedValue.
// V contains the bare value (not wrapped with @type tags).
// Nested struct fields and complex array elements are wrapped with JSONTypedValue.
type JSONTypedValue struct {
	T         string  // Type string: [<pkgpath>.]<symbol> or primitive
	V         any     // Value: null when nil, omitted only when Truncated=true
	ObjectID  *string // For pointers - enables later fetching
	Error     *string // .Error() result if implements error
	Base      *string // baseOf(T) if different from T
	Truncated bool    // True when content omitted due to depth limit
}

// MarshalJSON implements custom JSON marshaling for JSONTypedValue.
// V is omitted only when Truncated is true; otherwise V is included (even if nil -> null).
func (jtv JSONTypedValue) MarshalJSON() ([]byte, error) {
	type jsonAlias struct {
		T         string  `json:"T"`
		V         any     `json:"V,omitempty"`
		ObjectID  *string `json:"objectid,omitempty"`
		Error     *string `json:"error,omitempty"`
		Base      *string `json:"base,omitempty"`
		Truncated bool    `json:"truncated,omitempty"`
	}

	if jtv.Truncated {
		// Omit V when truncated
		return json.Marshal(jsonAlias{
			T:         jtv.T,
			ObjectID:  jtv.ObjectID,
			Error:     jtv.Error,
			Base:      jtv.Base,
			Truncated: jtv.Truncated,
		})
	}

	// Include V normally (even if nil -> null)
	type jsonAliasWithV struct {
		T        string  `json:"T"`
		V        any     `json:"V"`
		ObjectID *string `json:"objectid,omitempty"`
		Error    *string `json:"error,omitempty"`
		Base     *string `json:"base,omitempty"`
	}
	return json.Marshal(jsonAliasWithV{
		T:        jtv.T,
		V:        jtv.V,
		ObjectID: jtv.ObjectID,
		Error:    jtv.Error,
		Base:     jtv.Base,
	})
}

// JSONExportObject exports an Object to JSON using Amino marshaling.
// This function handles all object types including HeapItemValue, StructValue,
// ArrayValue, MapValue.
//
// The export process:
// 1. Root object is always fully expanded
// 2. HeapItemValue is automatically unwrapped to show the underlying value
//    (HeapItemValue is an implementation detail for pointer indirection)
// 3. Nested real objects (persisted) become RefValue with their actual ObjectID
//    (these ObjectIDs can be queried via vm/qobject)
// 4. Nested non-real objects are expanded inline (they have no queryable ObjectID)
// 5. Cycles are detected and converted to RefValue to prevent infinite recursion
//
// Then the exported object is serialized via Amino's MarshalJSONAny which
// produces @type tags like "/gno.StructValue" with full ObjectInfo.
func JSONExportObject(m *Machine, obj Object, maxDepth int) ([]byte, error) {
	seen := map[Object]int{}

	var st Store
	if m != nil {
		st = m.Store
	}

	// Unwrap HeapItemValue - it's an implementation detail that users shouldn't see.
	// When querying a HeapItemValue, show the underlying value instead.
	obj = unwrapHeapItemValue(st, obj)

	// Export the object using JSON-specific export that:
	// - Expands root and non-real objects inline
	// - Converts real nested objects to RefValue with their actual ObjectID
	exported := jsonExportObjectValue(st, obj, seen, 0, maxDepth)

	// Now serialize the exported (cycle-safe) object via Amino.
	return amino.MarshalJSONAny(exported)
}

// unwrapHeapItemValue unwraps a HeapItemValue to reveal the underlying object.
// If the inner value is a RefValue, it loads the actual object from the store.
// Returns the original object if it's not a HeapItemValue or can't be unwrapped.
func unwrapHeapItemValue(st Store, obj Object) Object {
	hiv, ok := obj.(*HeapItemValue)
	if !ok {
		return obj
	}

	// If the underlying value is directly an Object, use that
	if innerObj, ok := hiv.Value.V.(Object); ok {
		return innerObj
	}

	// If the underlying value is a RefValue, load the actual object from store
	if rv, ok := hiv.Value.V.(RefValue); ok && st != nil {
		if innerObj := st.GetObject(rv.ObjectID); innerObj != nil {
			return innerObj
		}
	}

	// Can't unwrap - return original HeapItemValue
	return obj
}

// jsonExportObjectValue exports an Object for JSON serialization.
// Unlike exportCopyValueWithRefs (for persistence), this function:
// - Always expands the root object (depth 0)
// - Unwraps HeapItemValue to show the underlying object directly
// - Expands non-real objects inline (they can't be queried separately)
// - Converts real nested objects to RefValue with actual ObjectID (queryable)
func jsonExportObjectValue(st Store, obj Object, seen map[Object]int, depth, maxDepth int) Value {
	if obj == nil {
		return nil
	}

	// Unwrap HeapItemValue - it's an implementation detail.
	// For nested real HeapItemValues, return RefValue pointing to the HeapItemValue
	// (since that's what qobject queries - and it will auto-unwrap there too).
	if hiv, ok := obj.(*HeapItemValue); ok {
		if depth > 0 && hiv.GetIsReal() {
			return RefValue{
				ObjectID: hiv.GetObjectID(),
				Escaped:  hiv.GetIsEscaped() || hiv.GetIsNewEscaped(),
				Hash:     hiv.GetHash(),
			}
		}
		// For root or non-real HeapItemValue, unwrap and export the inner object
		obj = unwrapHeapItemValue(st, obj)
	}

	// Check for cycles first
	if id, exists := seen[obj]; exists {
		// Cycle detected - return RefValue
		// Use actual ObjectID if real, otherwise use local ID
		if obj.GetIsReal() {
			return RefValue{
				ObjectID: obj.GetObjectID(),
				Escaped:  true,
			}
		}
		return RefValue{
			ObjectID: ObjectID{NewTime: uint64(id)},
			Escaped:  true,
		}
	}

	// For non-root real objects, return RefValue with actual ObjectID
	// This allows the user to query them via vm/qobject
	if depth > 0 && obj.GetIsReal() {
		return RefValue{
			ObjectID: obj.GetObjectID(),
			Escaped:  obj.GetIsEscaped() || obj.GetIsNewEscaped(),
			Hash:     obj.GetHash(),
		}
	}

	// Check depth limit (only for non-root)
	if maxDepth >= 0 && depth > maxDepth {
		if obj.GetIsReal() {
			return RefValue{
				ObjectID: obj.GetObjectID(),
				Escaped:  true,
			}
		}
		// Non-real object at depth limit - assign local ID
		id := len(seen) + 1
		seen[obj] = id
		return RefValue{
			ObjectID: ObjectID{NewTime: uint64(id)},
			Escaped:  true,
		}
	}

	// Mark as seen before recursing (cycle detection)
	id := len(seen) + 1
	seen[obj] = id

	// Expand the object
	return jsonExportCopyValue(st, obj, seen, depth, maxDepth)
}

// jsonExportCopyValue creates an exported copy of a Value for JSON serialization.
// This is similar to exportCopyValueWithRefs but uses jsonExportObjectValue for nested objects.
func jsonExportCopyValue(st Store, val Value, seen map[Object]int, depth, maxDepth int) Value {
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
		// Base can be an Object or RefValue (if already exported/persisted)
		var base Value
		switch b := cv.Base.(type) {
		case Object:
			base = jsonExportObjectValue(st, b, seen, depth+1, maxDepth)
		case RefValue:
			// Already a RefValue - keep as-is (it's a reference to a persisted object)
			base = b
		default:
			panic(fmt.Sprintf("unexpected pointer base type: %T", cv.Base))
		}
		return PointerValue{
			Base:  base,
			Index: cv.Index,
		}
	case *ArrayValue:
		if cv.Data == nil {
			list := make([]TypedValue, len(cv.List))
			for i, etv := range cv.List {
				list[i] = jsonExportTypedValue(st, etv, seen, depth+1, maxDepth)
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
		var base Value
		if cv.Base != nil {
			if obj, ok := cv.Base.(Object); ok {
				base = jsonExportObjectValue(st, obj, seen, depth+1, maxDepth)
			} else if ref, ok := cv.Base.(RefValue); ok {
				base = ref
			}
		}
		return &SliceValue{
			Base:   base,
			Offset: cv.Offset,
			Length: cv.Length,
			Maxcap: cv.Maxcap,
		}
	case *StructValue:
		fields := make([]TypedValue, len(cv.Fields))
		for i, ftv := range cv.Fields {
			fields[i] = jsonExportTypedValue(st, ftv, seen, depth+1, maxDepth)
		}
		return &StructValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Fields:     fields,
		}
	case *FuncValue:
		source := toRefNode(cv.Source)
		var parent Value
		if cv.Parent != nil {
			if obj, ok := cv.Parent.(Object); ok {
				parent = jsonExportObjectValue(st, obj, seen, depth+1, maxDepth)
			} else if ref, ok := cv.Parent.(RefValue); ok {
				parent = ref
			}
		}
		captures := make([]TypedValue, len(cv.Captures))
		for i, ctv := range cv.Captures {
			captures[i] = jsonExportTypedValue(st, ctv, seen, depth+1, maxDepth)
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
		fnc := jsonExportCopyValue(st, cv.Func, seen, depth, maxDepth).(*FuncValue)
		rtv := jsonExportTypedValue(st, cv.Receiver, seen, depth+1, maxDepth)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			key2 := jsonExportTypedValue(st, cur.Key, seen, depth+1, maxDepth)
			val2 := jsonExportTypedValue(st, cur.Value, seen, depth+1, maxDepth)
			list.Append(nilAllocator, key2).Value = val2
		}
		return &MapValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			List:       list,
		}
	case TypeValue:
		return toTypeValue(exportCopyTypeWithRefs(cv.Type, seen))
	case *PackageValue:
		// Packages always become RefValue
		return RefValue{
			PkgPath: cv.PkgPath,
		}
	case *Block:
		source := toRefNode(cv.Source)
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = jsonExportTypedValue(st, tv, seen, depth+1, maxDepth)
		}
		var bparent Value
		if cv.Parent != nil {
			if obj, ok := cv.Parent.(Object); ok {
				bparent = jsonExportObjectValue(st, obj, seen, depth+1, maxDepth)
			} else if ref, ok := cv.Parent.(RefValue); ok {
				bparent = ref
			}
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
			Value:      jsonExportTypedValue(st, cv.Value, seen, depth+1, maxDepth),
		}
	default:
		panic(fmt.Sprintf("unexpected type %v", reflect.TypeOf(val)))
	}
}

// jsonExportTypedValue exports a TypedValue for JSON serialization.
func jsonExportTypedValue(st Store, tv TypedValue, seen map[Object]int, depth, maxDepth int) TypedValue {
	result := TypedValue{}
	if tv.T != nil {
		result.T = exportRefOrCopyType(tv.T, seen)
	}

	if obj, ok := tv.V.(Object); ok {
		result.V = jsonExportObjectValue(st, obj, seen, depth, maxDepth)
		return result
	}

	// For non-Object values, use the JSON export copy
	if tv.V != nil {
		result.V = jsonExportCopyValue(st, tv.V, seen, depth, maxDepth)
	}
	// Copy primitive values
	result.N = tv.N
	return result
}

// JSONExportTypedValuesSimple produces human-readable JSON for TypedValues.
// Unlike JSONExportTypedValues, it outputs values directly without @type tags or ObjectInfo.
// Requires Machine to call .Error() methods on error types.
// Uses DefaultMaxDepth for depth limiting.
func JSONExportTypedValuesSimple(m *Machine, tvs []TypedValue) ([]byte, error) {
	return JSONExportTypedValuesWithDepth(m, tvs, DefaultMaxDepth)
}

// JSONExportTypedValuesWithDepth exports TypedValues with explicit depth control.
// maxDepth: -1 = unlimited, 0 = type only, 1+ = levels to expand
func JSONExportTypedValuesWithDepth(m *Machine, tvs []TypedValue, maxDepth int) ([]byte, error) {
	seen := map[Object]int{}
	return jsonExportTypedValuesWithDepth(m, tvs, seen, 0, maxDepth)
}

// jsonExportTypedValuesWithDepth is the internal implementation with depth tracking.
func jsonExportTypedValuesWithDepth(m *Machine, tvs []TypedValue, seen map[Object]int, depth, maxDepth int) ([]byte, error) {
	results := make([]*JSONTypedValue, len(tvs))
	for i, tv := range tvs {
		results[i] = jsonTypedValueSimpleWithDepth(m, tv, seen, depth, maxDepth)
	}
	return json.Marshal(results)
}

// jsonTypedValueSimple creates a JSONTypedValue from a TypedValue.
// Deprecated: Use jsonTypedValueSimpleWithDepth instead.
func jsonTypedValueSimple(m *Machine, tv TypedValue, seen map[Object]int) *JSONTypedValue {
	return jsonTypedValueSimpleWithDepth(m, tv, seen, 0, UnlimitedDepth)
}

// jsonTypedValueSimpleWithDepth creates a JSONTypedValue with depth limiting.
func jsonTypedValueSimpleWithDepth(m *Machine, tv TypedValue, seen map[Object]int, depth, maxDepth int) *JSONTypedValue {
	jtv := &JSONTypedValue{
		T: simpleTypeString(tv.T),
	}

	// Check for real object boundary - stop expansion with @ref
	// Skip this check at depth 0 (root level) to always show the queried object's content
	if depth > 0 {
		if obj, isObj := tv.V.(Object); isObj && obj.GetIsReal() {
			oid := obj.GetObjectID().String()
			jtv.V = map[string]string{"@ref": oid}
			jtv.ObjectID = &oid
			return jtv
		}
	}

	// Check depth limit (primitives always expand)
	bt := BaseOf(tv.T)
	_, isPrim := bt.(PrimitiveType)
	if !isPrim && maxDepth >= 0 && depth > maxDepth {
		// At depth limit - mark as truncated, omit V
		jtv.Truncated = true
		// Still include objectid for pointers if available
		if pv, ok := tv.V.(PointerValue); ok && pv.Base != nil {
			if obj, ok := pv.Base.(Object); ok {
				oid := obj.GetObjectID().String()
				if oid != ":0" {
					jtv.ObjectID = &oid
				}
			}
		}
		return jtv
	}

	// Expand value normally
	jtv.V = jsonValueSimpleWithDepth(m, tv, seen, depth, maxDepth)

	// For pointers, include ObjectID for later fetching
	if _, isPtr := BaseOf(tv.T).(*PointerType); isPtr && tv.V != nil {
		if pv, ok := tv.V.(PointerValue); ok {
			if base := pv.Base; base != nil {
				if obj, ok := base.(Object); ok {
					oid := obj.GetObjectID().String()
					if oid != ":0" { // Don't include empty ObjectID
						jtv.ObjectID = &oid
					}
				}
			}
		}
	}

	// For error types, include error string
	if tv.V != nil && tv.T != nil {
		if errStr := getSimpleErrorString(m, tv); errStr != nil {
			jtv.Error = errStr
		}
	}

	// For declared types, include base type string
	jtv.Base = simpleBaseTypeString(tv.T)

	return jtv
}

// simpleTypeString returns a human-readable type string.
func simpleTypeString(t Type) string {
	if t == nil {
		return "nil"
	}
	switch ct := t.(type) {
	case PrimitiveType:
		return ct.String()
	case *DeclaredType:
		return ct.TypeID().String()
	case *PointerType:
		return "*" + simpleTypeString(ct.Elt)
	case *SliceType:
		return "[]" + simpleTypeString(ct.Elt)
	case *ArrayType:
		return fmt.Sprintf("[%d]%s", ct.Len, simpleTypeString(ct.Elt))
	case *MapType:
		return fmt.Sprintf("map[%s]%s", simpleTypeString(ct.Key), simpleTypeString(ct.Value))
	case *FuncType:
		return "func"
	case *InterfaceType:
		if ct.IsEmptyInterface() {
			return "interface{}"
		}
		return "interface"
	case *StructType:
		return "struct"
	case RefType:
		return ct.ID.String()
	default:
		return t.String()
	}
}

// simpleBaseTypeString returns the base type string if it differs from the declared type.
func simpleBaseTypeString(t Type) *string {
	if _, ok := t.(*DeclaredType); ok {
		base := BaseOf(t)
		if base != nil && base != t {
			s := simpleTypeString(base)
			// Only include base if it's different and meaningful
			if s != "" && s != "struct" {
				return &s
			}
		}
	}
	return nil
}

// getJSONFieldName returns the JSON field name for a struct field.
// If a json tag is present, uses the tag name (without options like omitempty).
// Otherwise, returns the Go field name.
func getJSONFieldName(ft FieldType) string {
	if ft.Tag != "" {
		// Use reflect.StructTag to parse the tag
		tag := reflect.StructTag(ft.Tag)
		if jsonTag := tag.Get("json"); jsonTag != "" {
			// Split to handle "name,omitempty" -> take just "name"
			if comma := strings.Index(jsonTag, ","); comma != -1 {
				jsonTag = jsonTag[:comma]
			}
			// "-" means skip this field, but we still include it with Go name
			if jsonTag != "-" && jsonTag != "" {
				return jsonTag
			}
		}
	}
	return string(ft.Name)
}

// jsonValueSimple converts a TypedValue's value to a JSON-compatible representation.
// Deprecated: Use jsonValueSimpleWithDepth instead.
func jsonValueSimple(m *Machine, tv TypedValue, seen map[Object]int) any {
	return jsonValueSimpleWithDepth(m, tv, seen, 0, UnlimitedDepth)
}

// jsonValueSimpleWithDepth converts a TypedValue's value to JSON with depth limiting.
func jsonValueSimpleWithDepth(m *Machine, tv TypedValue, seen map[Object]int, depth, maxDepth int) any {
	if tv.T == nil {
		return nil
	}

	bt := BaseOf(tv.T)

	switch bt := bt.(type) {
	case PrimitiveType:
		// Primitives always expand regardless of depth
		return getPrimitiveValueSimple(bt, tv)

	case *StructType:
		sv, ok := tv.V.(*StructValue)
		if !ok {
			return nil
		}

		// Check for real object boundary - stop expansion with @ref
		// Skip this check at depth 0 (root level) to always show the queried object's content
		if depth > 0 && sv.GetIsReal() {
			oid := sv.GetObjectID().String()
			return map[string]string{"@ref": oid}
		}

		// Cycle check
		if id, exists := seen[sv]; exists {
			return map[string]string{"@ref": fmt.Sprintf(":%d", id)}
		}
		id := len(seen) + 1
		seen[sv] = id

		// Use map[string]any to allow mixing objectid with field values
		obj := make(map[string]any)

		// Add objectid for unreal objects to enable cycle reference tracking
		obj["objectid"] = fmt.Sprintf(":%d", id)

		for i := range sv.Fields {
			if i < len(bt.Fields) {
				name := getJSONFieldName(bt.Fields[i])
				// Fill the field value to resolve any RefValues before serialization
				field := fillValueTV(m.Store, &sv.Fields[i])
				obj[name] = jsonTypedValueSimpleWithDepth(m, *field, seen, depth+1, maxDepth)
			}
		}
		return obj

	case *SliceType:
		return jsonSliceValueSimpleWithDepth(m, tv, bt, seen, depth, maxDepth)

	case *ArrayType:
		return jsonArrayValueSimpleWithDepth(m, tv, bt, seen, depth, maxDepth)

	case *PointerType:
		pv, ok := tv.V.(PointerValue)
		if !ok {
			return nil
		}
		// Handle case where TV is nil
		if pv.TV == nil {
			// Helper to get TV from base
			getFromBase := func(base Value) any {
				switch cbv := base.(type) {
				case *ArrayValue:
					et := bt.Elt
					epv := cbv.GetPointerAtIndexInt2(m.Store, pv.Index, et)
					if epv.TV == nil {
						return nil
					}
					return jsonValueSimpleWithDepth(m, *epv.TV, seen, depth, maxDepth)
				case *StructValue:
					fpv := cbv.GetPointerToInt(m.Store, pv.Index)
					if fpv.TV == nil {
						return nil
					}
					return jsonValueSimpleWithDepth(m, *fpv.TV, seen, depth, maxDepth)
				case *Block:
					vpv := cbv.GetPointerToInt(m.Store, pv.Index)
					if vpv.TV == nil {
						return nil
					}
					return jsonValueSimpleWithDepth(m, *vpv.TV, seen, depth, maxDepth)
				case *HeapItemValue:
					return jsonValueSimpleWithDepth(m, cbv.Value, seen, depth, maxDepth)
				default:
					return nil
				}
			}

			// Check if Base is RefValue (needs loading from store)
			if rv, isRef := pv.Base.(RefValue); isRef && m.Store != nil {
				base := m.Store.GetObject(rv.ObjectID)
				if base == nil {
					return nil
				}
				return getFromBase(base.(Value))
			}
			// Base is already filled (not RefValue) - use it directly
			if pv.Base != nil {
				return getFromBase(pv.Base)
			}
			return nil
		}
		// Dereference - but the TV might contain a RefValue that needs filling
		deref := pv.Deref()
		// Fill the dereferenced value to resolve any nested RefValues
		if m.Store != nil {
			fillValueTV(m.Store, &deref)
		}
		return jsonValueSimpleWithDepth(m, deref, seen, depth, maxDepth)

	case *MapType:
		return jsonMapValueSimpleWithDepth(m, tv, bt, seen, depth, maxDepth)

	case *FuncType:
		if fv, ok := tv.V.(*FuncValue); ok {
			if fv.PkgPath != "" {
				return fv.PkgPath + "." + string(fv.Name)
			}
			return string(fv.Name)
		}
		return "<func>"

	case *InterfaceType:
		// For interface types, serialize the concrete value
		return jsonValueSimpleWithDepth(m, tv, seen, depth, maxDepth)

	default:
		// Fallback: return string representation
		if tv.V != nil {
			return tv.V.String()
		}
		return nil
	}
}

// getPrimitiveValueSimple returns the Go value for a primitive type.
func getPrimitiveValueSimple(bt PrimitiveType, tv TypedValue) any {
	switch bt {
	case UntypedBoolType, BoolType:
		return tv.GetBool()
	case UntypedStringType, StringType:
		return tv.GetString()
	case IntType:
		return tv.GetInt()
	case Int8Type:
		return int64(tv.GetInt8())
	case Int16Type:
		return int64(tv.GetInt16())
	case UntypedRuneType, Int32Type:
		return int64(tv.GetInt32())
	case Int64Type:
		return tv.GetInt64()
	case UintType:
		return tv.GetUint()
	case Uint8Type:
		return uint64(tv.GetUint8())
	case DataByteType:
		return uint64(tv.GetDataByte())
	case Uint16Type:
		return uint64(tv.GetUint16())
	case Uint32Type:
		return uint64(tv.GetUint32())
	case Uint64Type:
		return tv.GetUint64()
	case Float32Type:
		return math.Float32frombits(tv.GetFloat32())
	case Float64Type:
		return math.Float64frombits(tv.GetFloat64())
	case UntypedBigintType:
		return tv.V.(BigintValue).V.String()
	case UntypedBigdecType:
		return tv.V.(BigdecValue).V.String()
	default:
		return nil
	}
}

// jsonSliceValueSimple converts a slice value to JSON.
// Deprecated: Use jsonSliceValueSimpleWithDepth instead.
func jsonSliceValueSimple(m *Machine, tv TypedValue, st *SliceType, seen map[Object]int) any {
	return jsonSliceValueSimpleWithDepth(m, tv, st, seen, 0, UnlimitedDepth)
}

// jsonSliceValueSimpleWithDepth converts a slice value to JSON with depth limiting.
func jsonSliceValueSimpleWithDepth(m *Machine, tv TypedValue, st *SliceType, seen map[Object]int, depth, maxDepth int) any {
	sv, ok := tv.V.(*SliceValue)
	if !ok {
		return nil
	}

	length := sv.Length
	if length == 0 {
		return []any{}
	}

	// Check if element type is primitive
	if isPrimitiveType(st.Elt) {
		// Return direct array for primitives (no depth limit for primitives)
		result := make([]any, length)
		for i := range length {
			etv := sv.GetPointerAtIndexInt2(m.Store, i, st.Elt).Deref()
			result[i] = getPrimitiveValueSimple(BaseOf(st.Elt).(PrimitiveType), etv)
		}
		return result
	}

	// For complex types, return array of JSONTypedValue
	result := make([]*JSONTypedValue, length)
	for i := range length {
		etv := sv.GetPointerAtIndexInt2(m.Store, i, st.Elt).Deref()
		result[i] = jsonTypedValueSimpleWithDepth(m, etv, seen, depth+1, maxDepth)
	}
	return result
}

// jsonArrayValueSimple converts an array value to JSON.
// Deprecated: Use jsonArrayValueSimpleWithDepth instead.
func jsonArrayValueSimple(m *Machine, tv TypedValue, at *ArrayType, seen map[Object]int) any {
	return jsonArrayValueSimpleWithDepth(m, tv, at, seen, 0, UnlimitedDepth)
}

// jsonArrayValueSimpleWithDepth converts an array value to JSON with depth limiting.
func jsonArrayValueSimpleWithDepth(m *Machine, tv TypedValue, at *ArrayType, seen map[Object]int, depth, maxDepth int) any {
	av, ok := tv.V.(*ArrayValue)
	if !ok {
		return nil
	}

	length := at.Len
	if length == 0 {
		return []any{}
	}

	// Handle byte arrays specially (Data field) - no depth limit for byte arrays
	if av.Data != nil {
		// Return as array of numbers
		result := make([]any, length)
		for i := 0; i < length; i++ {
			result[i] = uint64(av.Data[i])
		}
		return result
	}

	// Check if element type is primitive (no depth limit for primitives)
	if isPrimitiveType(at.Elt) {
		result := make([]any, length)
		for i := 0; i < length; i++ {
			result[i] = getPrimitiveValueSimple(BaseOf(at.Elt).(PrimitiveType), av.List[i])
		}
		return result
	}

	// For complex types, return array of JSONTypedValue
	result := make([]*JSONTypedValue, length)
	for i := range length {
		result[i] = jsonTypedValueSimpleWithDepth(m, av.List[i], seen, depth+1, maxDepth)
	}
	return result
}

// jsonMapValueSimple converts a map value to JSON.
// Deprecated: Use jsonMapValueSimpleWithDepth instead.
func jsonMapValueSimple(m *Machine, tv TypedValue, mt *MapType, seen map[Object]int) any {
	return jsonMapValueSimpleWithDepth(m, tv, mt, seen, 0, UnlimitedDepth)
}

// jsonMapValueSimpleWithDepth converts a map value to JSON with depth limiting.
func jsonMapValueSimpleWithDepth(m *Machine, tv TypedValue, mt *MapType, seen map[Object]int, depth, maxDepth int) any {
	mv, ok := tv.V.(*MapValue)
	if !ok {
		return nil
	}

	// Check if key type is string-like
	if isStringLikeType(mt.Key) {
		// Return as JSON object
		obj := make(map[string]*JSONTypedValue)
		for cur := mv.List.Head; cur != nil; cur = cur.Next {
			keyStr := fmt.Sprintf("%v", jsonValueSimpleWithDepth(m, cur.Key, seen, depth, maxDepth))
			obj[keyStr] = jsonTypedValueSimpleWithDepth(m, cur.Value, seen, depth+1, maxDepth)
		}
		return obj
	}

	// Return as array of [key, value] pairs
	var pairs [][]any
	for cur := mv.List.Head; cur != nil; cur = cur.Next {
		pair := []any{
			jsonTypedValueSimpleWithDepth(m, cur.Key, seen, depth+1, maxDepth),
			jsonTypedValueSimpleWithDepth(m, cur.Value, seen, depth+1, maxDepth),
		}
		pairs = append(pairs, pair)
	}
	return pairs
}

// isPrimitiveType returns true if the type is a primitive type.
func isPrimitiveType(t Type) bool {
	_, ok := BaseOf(t).(PrimitiveType)
	return ok
}

// isStringLikeType returns true if the type can be used as a JSON object key.
func isStringLikeType(t Type) bool {
	bt := BaseOf(t)
	if pt, ok := bt.(PrimitiveType); ok {
		return pt == StringType || pt == UntypedStringType
	}
	return false
}

// ExportCopyObjectWithRefs exports an Object with references replaced by RefValue.
// This is the public API for QueryObject to export objects retrieved by ObjectID.
func ExportCopyObjectWithRefs(obj Object, seen map[Object]int) Value {
	return exportCopyValueWithRefs(obj, seen)
}

// getSimpleErrorString attempts to get the error string from a TypedValue.
func getSimpleErrorString(m *Machine, tv TypedValue) *string {
	// Check if value implements error
	if !tv.ImplError() {
		return nil
	}

	// Try to call .Error() method
	defer func() {
		recover() // Ignore panics from method calls
	}()

	// Wrap in pointer if needed for method call
	callTV := tv
	if _, ok := BaseOf(tv.T).(*PointerType); !ok {
		callTV = TypedValue{
			T: &PointerType{Elt: tv.T},
			V: PointerValue{TV: &tv},
		}
	}

	res := m.Eval(Call(Sel(&ConstExpr{TypedValue: callTV}, "Error")))
	if len(res) > 0 {
		s := res[0].GetString()
		return &s
	}
	return nil
}

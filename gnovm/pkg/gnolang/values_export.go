package gnolang

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// JSONExporter handles JSON serialization of Gno values.
// Configuration fields control export behavior.
type JSONExporter struct {
	// ExportUnexported controls whether unexported (lowercase) fields are included.
	// Default is false (only export uppercase fields).
	ExportUnexported bool

	// MaxDepth limits the depth of nested object expansion.
	// Default is DefaultMaxDepth (3). Use -1 for unlimited.
	MaxDepth int

	// Internal state (set during export)
	store Store
	seen  map[Object]int
}

// NewJSONExporter creates an exporter with default options.
func NewJSONExporter() *JSONExporter {
	return &JSONExporter{
		ExportUnexported: false,
		MaxDepth:         DefaultMaxDepth,
	}
}

// init initializes internal state for a new export operation.
func (e *JSONExporter) init() {
	if e.seen == nil {
		e.seen = map[Object]int{}
	}
	if e.MaxDepth == 0 {
		e.MaxDepth = DefaultMaxDepth
	}
}

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

func exportToValueOrRefValue(val Value, seen map[Object]int) Value {
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

// ExportTypedValues exports multiple TypedValues to JSON.
// Persisted objects are shown as RefValue (with queryable ObjectID).
// Ephemeral objects are expanded inline.
func (e *JSONExporter) ExportTypedValues(tvs []TypedValue) ([]byte, error) {
	e.init()

	jexps := make([]*jsonTypedValue, len(tvs))

	for i, tv := range tvs {
		// Use JSON-specific export that:
		// - Shows RefValue for persisted (real) objects (with queryable ObjectID)
		// - Expands ephemeral (unreal) objects inline (no ObjectID)
		// - Filters unexported fields and json:"-" tagged fields based on options
		exported := e.exportTypedValue(tv, 0)
		jexps[i] = jsonExportedTypedValueFromExported(exported)
	}

	return json.Marshal(jexps)
}

// JSONExportTypedValues exports TypedValues using default options.
// For custom options, use JSONExporter directly.
func JSONExportTypedValues(tvs []TypedValue) ([]byte, error) {
	return NewJSONExporter().ExportTypedValues(tvs)
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

// jsonExportedTypedValueFromExported converts an already-exported TypedValue to JSON format.
// This is used by JSONExportTypedValues() after calling jsonExportTypedValue() which
// expands ephemeral objects inline while keeping RefValue for persisted objects.
func jsonExportedTypedValueFromExported(tv TypedValue) *jsonTypedValue {
	return &jsonTypedValue{
		Type:  jsonExportedType(tv.T),
		Value: jsonExportedValueFromExported(tv),
	}
}

// jsonExportedValueFromExported converts an already-exported TypedValue's value to JSON.
// Unlike jsonExportedValue(), this handles TypedValues that have already been processed
// by jsonExportTypedValue() - meaning primitive N values have NOT been extracted yet.
func jsonExportedValueFromExported(tv TypedValue) []byte {
	bt := BaseOf(tv.T)
	switch bt := bt.(type) {
	case PrimitiveType:
		var ret string
		switch bt {
		case UntypedBoolType, BoolType:
			ret = strconv.FormatBool(tv.GetBool())
		case UntypedStringType, StringType:
			if sv, ok := tv.V.(StringValue); ok {
				ret = strconv.Quote(string(sv))
			} else {
				ret = strconv.Quote(tv.GetString())
			}
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
	DefaultMaxDepth = 3 // Default depth limit for JSON export
)

// isExportedName returns true if name starts with an uppercase letter.
// This follows Go's convention where uppercase names are exported (public).
func isExportedName(name Name) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

// hasJSONSkipTag returns true if the struct tag contains `json:"-"`.
// Fields with this tag should be omitted from JSON output.
func hasJSONSkipTag(tag Tag) bool {
	if tag == "" {
		return false
	}
	// Parse the json tag from the struct tag
	// Format: `json:"name,options"` or `json:"-"`
	tagStr := string(tag)
	for tagStr != "" {
		// Skip leading space
		i := 0
		for i < len(tagStr) && tagStr[i] == ' ' {
			i++
		}
		tagStr = tagStr[i:]
		if tagStr == "" {
			break
		}

		// Scan to colon - find the key
		i = 0
		for i < len(tagStr) && tagStr[i] > ' ' && tagStr[i] != ':' && tagStr[i] != '"' && tagStr[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tagStr) || tagStr[i] != ':' || tagStr[i+1] != '"' {
			break
		}
		key := tagStr[:i]
		tagStr = tagStr[i+1:]

		// Scan quoted value
		i = 1
		for i < len(tagStr) && tagStr[i] != '"' {
			if tagStr[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tagStr) {
			break
		}
		value := tagStr[1:i]
		tagStr = tagStr[i+1:]

		if key == "json" {
			// Check if the json tag value is "-" or starts with "-,"
			if value == "-" || strings.HasPrefix(value, "-,") {
				return true
			}
			return false
		}
	}
	return false
}

// getStructTypeFromType resolves a Type to its underlying *StructType.
// Returns nil if the type is not a struct type.
func getStructTypeFromType(typ Type) *StructType {
	if typ == nil {
		return nil
	}

	switch t := typ.(type) {
	case *DeclaredType:
		return getStructTypeFromType(t.Base)
	case *StructType:
		return t
	case RefType:
		// RefType needs to be resolved via store, which we don't have here.
		// Caller should resolve RefType before calling this function.
		return nil
	default:
		return nil
	}
}

// ExportObject exports an Object to JSON using Amino marshaling.
// This function handles all object types including HeapItemValue, StructValue,
// ArrayValue, MapValue.
//
// The export process:
// 1. Root object is always fully expanded
// 2. HeapItemValue is automatically unwrapped to show the underlying value
//    (HeapItemValue is an implementation detail for pointer indirection)
// 3. Nested real objects (persisted) become RefValue with their actual ObjectID
// 4. Nested non-real objects are expanded inline (no queryable ObjectID)
// 5. Cycles are detected and converted to RefValue to prevent infinite recursion
//
// The exported object is serialized via Amino's MarshalJSONAny which
// produces @type tags like "/gno.StructValue" with full ObjectInfo.
func (e *JSONExporter) ExportObject(m *Machine, obj Object) ([]byte, error) {
	e.init()

	if m != nil {
		e.store = m.Store
	}

	// Unwrap HeapItemValue - it's an implementation detail that users shouldn't see.
	// When querying a HeapItemValue, show the underlying value instead.
	obj = unwrapHeapItemValue(e.store, obj)

	// Export the object using JSON-specific export that:
	// - Expands root and non-real objects inline
	// - Converts real nested objects to RefValue with their actual ObjectID
	exported := jsonExportObjectValue(e.store, obj, e.seen, 0, e.MaxDepth)

	// Now serialize the exported (cycle-safe) object via Amino.
	return amino.MarshalJSONAny(exported)
}

// JSONExportObject exports an Object using default options.
// For custom options, use JSONExporter directly.
func JSONExportObject(m *Machine, obj Object) ([]byte, error) {
	return NewJSONExporter().ExportObject(m, obj)
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
	// For nested real HeapItemValues, return RefValue pointing to the HeapItemValue.
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

	// For non-root real objects, return RefValue with actual ObjectID.
	// The ObjectID can be used to fetch the full object separately.
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

// exportTypedValue exports a TypedValue for JSON serialization.
// Returns RefValue for persisted (real) objects at all levels.
// Only ephemeral (unreal) objects are expanded inline.
func (e *JSONExporter) exportTypedValue(tv TypedValue, depth int) TypedValue {
	result := TypedValue{}
	if tv.T != nil {
		result.T = exportRefOrCopyType(tv.T, e.seen)
	}

	if obj, ok := tv.V.(Object); ok {
		result.V = e.exportObjectValue(obj, tv.T, depth)
		return result
	}

	// For non-Object values, use the JSON export copy
	if tv.V != nil {
		result.V = e.exportCopyValue(tv.V, tv.T, depth)
	}
	// Copy primitive values
	result.N = tv.N
	return result
}

// exportObjectValue exports an Object for JSON serialization.
// Returns RefValue for persisted (real) objects at all levels (including root).
// The typ parameter provides type information for field filtering.
func (e *JSONExporter) exportObjectValue(obj Object, typ Type, depth int) Value {
	if obj == nil {
		return nil
	}

	// Unwrap HeapItemValue - it's an implementation detail.
	if hiv, ok := obj.(*HeapItemValue); ok {
		if hiv.GetIsReal() {
			// Real HeapItemValue - return RefValue pointing to it
			return RefValue{
				ObjectID: hiv.GetObjectID(),
				Escaped:  hiv.GetIsEscaped() || hiv.GetIsNewEscaped(),
				Hash:     hiv.GetHash(),
			}
		}
		// Non-real HeapItemValue - unwrap and continue
		obj = unwrapHeapItemValue(e.store, obj)
	}

	// Check for cycles first
	if id, exists := e.seen[obj]; exists {
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

	// For real objects at any level (including root), return RefValue.
	// The ObjectID can be used to fetch the full object separately.
	if obj.GetIsReal() {
		return RefValue{
			ObjectID: obj.GetObjectID(),
			Escaped:  obj.GetIsEscaped() || obj.GetIsNewEscaped(),
			Hash:     obj.GetHash(),
		}
	}

	// Check depth limit for non-real objects
	if e.MaxDepth >= 0 && depth > e.MaxDepth {
		id := len(e.seen) + 1
		e.seen[obj] = id
		return RefValue{
			ObjectID: ObjectID{NewTime: uint64(id)},
			Escaped:  true,
		}
	}

	// Mark as seen before recursing (cycle detection)
	id := len(e.seen) + 1
	e.seen[obj] = id

	// Expand non-real (ephemeral) object inline
	return e.exportCopyValue(obj, typ, depth)
}

// exportCopyValue creates an exported copy of a Value for JSON serialization.
// The typ parameter provides type information for field filtering.
func (e *JSONExporter) exportCopyValue(val Value, typ Type, depth int) Value {
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
		// Get element type for the pointer
		var eltType Type
		if pt, ok := BaseOf(typ).(*PointerType); ok {
			eltType = pt.Elt
		}
		var base Value
		switch b := cv.Base.(type) {
		case Object:
			base = e.exportObjectValue(b, eltType, depth+1)
		case RefValue:
			base = b
		default:
			panic(fmt.Sprintf("unexpected pointer base type: %T", cv.Base))
		}
		return PointerValue{
			Base:  base,
			Index: cv.Index,
		}
	case *ArrayValue:
		// Get element type for array
		var eltType Type
		if at, ok := BaseOf(typ).(*ArrayType); ok {
			eltType = at.Elt
		}
		if cv.Data == nil {
			list := make([]TypedValue, len(cv.List))
			for i, etv := range cv.List {
				etv.T = eltType // Set element type for filtering
				list[i] = e.exportTypedValue(etv, depth+1)
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
		// Get element type for slice
		var eltType Type
		if slt, ok := BaseOf(typ).(*SliceType); ok {
			eltType = slt.Elt
		}
		var base Value
		if cv.Base != nil {
			if obj, ok := cv.Base.(Object); ok {
				// Slice base is an array, create array type for it
				var arrType Type
				if eltType != nil {
					arrType = &ArrayType{Elt: eltType}
				}
				base = e.exportObjectValue(obj, arrType, depth+1)
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
		// Get struct type for field filtering
		structType := getStructTypeFromType(typ)

		if structType == nil || len(structType.Fields) != len(cv.Fields) {
			// No type info or mismatch - export all fields (fallback)
			fields := make([]TypedValue, len(cv.Fields))
			for i, ftv := range cv.Fields {
				fields[i] = e.exportTypedValue(ftv, depth+1)
			}
			return &StructValue{
				ObjectInfo: cv.ObjectInfo.Copy(),
				Fields:     fields,
			}
		}

		// Filter fields based on visibility and json tags
		var fields []TypedValue
		for i, ftv := range cv.Fields {
			ft := structType.Fields[i]

			// Always skip json:"-" tagged fields
			if hasJSONSkipTag(ft.Tag) {
				continue
			}

			// Check export visibility
			if !e.ExportUnexported && !isExportedName(ft.Name) {
				continue
			}

			// Export this field with its type
			ftv.T = ft.Type
			fields = append(fields, e.exportTypedValue(ftv, depth+1))
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
				parent = e.exportObjectValue(obj, nil, depth+1)
			} else if ref, ok := cv.Parent.(RefValue); ok {
				parent = ref
			}
		}
		captures := make([]TypedValue, len(cv.Captures))
		for i, ctv := range cv.Captures {
			captures[i] = e.exportTypedValue(ctv, depth+1)
		}
		ft := exportCopyTypeWithRefs(cv.Type, e.seen)
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
		fnc := e.exportCopyValue(cv.Func, nil, depth).(*FuncValue)
		rtv := e.exportTypedValue(cv.Receiver, depth+1)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		// Get key/value types from map type
		var keyType, valType Type
		if mt, ok := BaseOf(typ).(*MapType); ok {
			keyType = mt.Key
			valType = mt.Value
		}
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			k := cur.Key
			k.T = keyType
			v := cur.Value
			v.T = valType
			key2 := e.exportTypedValue(k, depth+1)
			val2 := e.exportTypedValue(v, depth+1)
			list.Append(nilAllocator, key2).Value = val2
		}
		return &MapValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			List:       list,
		}
	case TypeValue:
		return toTypeValue(exportCopyTypeWithRefs(cv.Type, e.seen))
	case *PackageValue:
		return RefValue{
			PkgPath: cv.PkgPath,
		}
	case *Block:
		source := toRefNode(cv.Source)
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = e.exportTypedValue(tv, depth+1)
		}
		var bparent Value
		if cv.Parent != nil {
			if obj, ok := cv.Parent.(Object); ok {
				bparent = e.exportObjectValue(obj, nil, depth+1)
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
			Value:      e.exportTypedValue(cv.Value, depth+1),
		}
	default:
		panic(fmt.Sprintf("unexpected type %v", reflect.TypeOf(val)))
	}
}

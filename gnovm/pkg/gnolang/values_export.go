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

// JSONTypedValue represents a human-readable JSON format for TypedValue.
// V contains the bare value (not wrapped with @type tags).
// Nested struct fields and complex array elements are wrapped with JSONTypedValue.
type JSONTypedValue struct {
	T        string  `json:"T"`                  // Type string: [<pkgpath>.]<symbol> or primitive
	V        any     `json:"V"`                  // Value: string | number | bool | null | object | array
	ObjectID *string `json:"objectid,omitempty"` // For pointers - enables later fetching
	Error    *string `json:"error,omitempty"`    // .Error() result if implements error
	Base     *string `json:"base,omitempty"`     // baseOf(T) if different from T
}

// JSONExportTypedValuesSimple produces human-readable JSON for TypedValues.
// Unlike JSONExportTypedValues, it outputs values directly without @type tags or ObjectInfo.
// Requires Machine to call .Error() methods on error types.
func JSONExportTypedValuesSimple(m *Machine, tvs []TypedValue) ([]byte, error) {
	seen := map[Object]int{}
	results := make([]*JSONTypedValue, len(tvs))
	for i, tv := range tvs {
		results[i] = jsonTypedValueSimple(m, tv, seen)
	}
	return json.Marshal(results)
}

// jsonTypedValueSimple creates a JSONTypedValue from a TypedValue.
func jsonTypedValueSimple(m *Machine, tv TypedValue, seen map[Object]int) *JSONTypedValue {
	jtv := &JSONTypedValue{
		T: simpleTypeString(tv.T),
		V: jsonValueSimple(m, tv, seen),
	}

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
func jsonValueSimple(m *Machine, tv TypedValue, seen map[Object]int) any {
	if tv.T == nil {
		return nil
	}

	bt := BaseOf(tv.T)

	switch bt := bt.(type) {
	case PrimitiveType:
		// Primitives store values in tv.N, not tv.V
		return getPrimitiveValueSimple(bt, tv)

	case *StructType:
		sv, ok := tv.V.(*StructValue)
		if !ok {
			return nil
		}
		// Cycle check
		if id, exists := seen[sv]; exists {
			return map[string]int{"@ref": id}
		}
		seen[sv] = len(seen) + 1

		// Each field is wrapped with JSONTypedValue
		obj := make(map[string]*JSONTypedValue)
		for i, field := range sv.Fields {
			if i < len(bt.Fields) {
				name := getJSONFieldName(bt.Fields[i])
				obj[name] = jsonTypedValueSimple(m, field, seen)
			}
		}
		return obj

	case *SliceType:
		return jsonSliceValueSimple(m, tv, bt, seen)

	case *ArrayType:
		return jsonArrayValueSimple(m, tv, bt, seen)

	case *PointerType:
		pv, ok := tv.V.(PointerValue)
		if !ok {
			return nil
		}
		// Dereference and serialize target value
		deref := pv.Deref()
		return jsonValueSimple(m, deref, seen)

	case *MapType:
		return jsonMapValueSimple(m, tv, bt, seen)

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
		return jsonValueSimple(m, tv, seen)

	default:
		// Fallback: return string representation
		return tv.V.String()
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
func jsonSliceValueSimple(m *Machine, tv TypedValue, st *SliceType, seen map[Object]int) any {
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
		// Return direct array for primitives
		result := make([]any, length)
		for i := 0; i < length; i++ {
			etv := sv.GetPointerAtIndexInt2(nil, i, st.Elt).Deref()
			result[i] = getPrimitiveValueSimple(BaseOf(st.Elt).(PrimitiveType), etv)
		}
		return result
	}

	// For complex types, return array of JSONTypedValue
	result := make([]*JSONTypedValue, length)
	for i := 0; i < length; i++ {
		etv := sv.GetPointerAtIndexInt2(nil, i, st.Elt).Deref()
		result[i] = jsonTypedValueSimple(m, etv, seen)
	}
	return result
}

// jsonArrayValueSimple converts an array value to JSON.
func jsonArrayValueSimple(m *Machine, tv TypedValue, at *ArrayType, seen map[Object]int) any {
	av, ok := tv.V.(*ArrayValue)
	if !ok {
		return nil
	}

	length := at.Len
	if length == 0 {
		return []any{}
	}

	// Handle byte arrays specially (Data field)
	if av.Data != nil {
		// Return as array of numbers
		result := make([]any, length)
		for i := 0; i < length; i++ {
			result[i] = uint64(av.Data[i])
		}
		return result
	}

	// Check if element type is primitive
	if isPrimitiveType(at.Elt) {
		result := make([]any, length)
		for i := 0; i < length; i++ {
			result[i] = getPrimitiveValueSimple(BaseOf(at.Elt).(PrimitiveType), av.List[i])
		}
		return result
	}

	// For complex types, return array of JSONTypedValue
	result := make([]*JSONTypedValue, length)
	for i := 0; i < length; i++ {
		result[i] = jsonTypedValueSimple(m, av.List[i], seen)
	}
	return result
}

// jsonMapValueSimple converts a map value to JSON.
func jsonMapValueSimple(m *Machine, tv TypedValue, mt *MapType, seen map[Object]int) any {
	mv, ok := tv.V.(*MapValue)
	if !ok {
		return nil
	}

	// Check if key type is string-like
	if isStringLikeType(mt.Key) {
		// Return as JSON object
		obj := make(map[string]*JSONTypedValue)
		for cur := mv.List.Head; cur != nil; cur = cur.Next {
			keyStr := fmt.Sprintf("%v", jsonValueSimple(m, cur.Key, seen))
			obj[keyStr] = jsonTypedValueSimple(m, cur.Value, seen)
		}
		return obj
	}

	// Return as array of [key, value] pairs
	var pairs [][]any
	for cur := mv.List.Head; cur != nil; cur = cur.Next {
		pair := []any{
			jsonTypedValueSimple(m, cur.Key, seen),
			jsonTypedValueSimple(m, cur.Value, seen),
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

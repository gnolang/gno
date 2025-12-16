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

type JSONExporterOptions struct {
	// ExportUnexported controls whether unexported (lowercase) fields are included.
	// Default is false (only export uppercase fields).
	ExportUnexported bool

	// MaxDepth limits the depth of nested object expansion.
	// 0 means no limit.
	MaxDepth int
}

// ExportTypedValues exports multiple TypedValues to JSON.
// Root values are expanded inline, nested real objects become RefValue.
func (opts JSONExporterOptions) ExportTypedValues(tvs []TypedValue) ([]byte, error) {
	e := &jsonExporter{opts: opts}
	e.init()

	jexps := make([]*jsonTypedValue, len(tvs))

	for i, tv := range tvs {
		// Export with RefValue for persisted objects, inline expansion for ephemeral objects.
		// Filters fields based on ExportUnexported option and json:"-" tags.
		exported := e.exportTypedValue(tv, 0)
		jexps[i] = e.jsonExportedTypedValueFromExported(exported)
	}

	return json.Marshal(jexps)
}

// exportCopyValueWithRefs copies a value with references to objects.
// This is used by exportCopyMethods for type serialization.
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
			Base:  exportToValueOrRefValue(cv.Base, m),
			Index: cv.Index,
		}
	case *ArrayValue:
		if cv.Data == nil {
			list := make([]TypedValue, len(cv.List))
			for i, etv := range cv.List {
				list[i] = exportTypedValueWithRefs(etv, m)
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
			fields[i] = exportTypedValueWithRefs(ftv, m)
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
			captures[i] = exportTypedValueWithRefs(ctv, m)
		}
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
		rtv := exportTypedValueWithRefs(cv.Receiver, m)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			key2 := exportTypedValueWithRefs(cur.Key, m)
			val2 := exportTypedValueWithRefs(cur.Value, m)
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
			FNames:     cv.FNames,
			FBlocks:    fblocks,
			Realm:      cv.Realm,
		}
	case *Block:
		source := toRefNode(cv.Source)
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = exportTypedValueWithRefs(tv, m)
		}
		var bparent Value
		if cv.Parent != nil {
			bparent = exportToValueOrRefValue(cv.Parent, m)
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
			Value:      exportTypedValueWithRefs(cv.Value, m),
		}
	default:
		panic(fmt.Sprintf("unexpected type %v", reflect.TypeOf(val)))
	}
}

// exportTypedValueWithRefs exports a TypedValue for type serialization.
func exportTypedValueWithRefs(tv TypedValue, m map[Object]int) TypedValue {
	result := TypedValue{N: tv.N}
	if tv.T != nil {
		result.T = exportRefOrCopyType(tv.T, m)
	}
	if obj, ok := tv.V.(Object); ok {
		result.V = exportToValueOrRefValue(obj, m)
		return result
	}
	if tv.V != nil {
		result.V = exportCopyValueWithRefs(tv.V, m)
	}
	return result
}

// exportToValueOrRefValue converts an object to RefValue if it's persisted,
// or copies it if it's ephemeral.
func exportToValueOrRefValue(val Value, seen map[Object]int) Value {
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
		return RefValue{PkgPath: pv.PkgPath}
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

	if id, ok := seen[oo]; ok {
		return RefValue{
			ObjectID: ObjectID{NewTime: uint64(id)},
			Escaped:  true,
		}
	}

	if oo.GetIsNewEscaped() {
		return RefValue{
			ObjectID: oo.GetObjectID(),
			Escaped:  true,
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

// jsonNamedField represents a struct field with its name for JSON export.
type jsonNamedField struct {
	Name  string          `json:"N"`
	Type  json.RawMessage `json:"T"`
	Value json.RawMessage `json:"V"`
}

// jsonStructValue is a StructValue with named fields for custom JSON marshaling.
// Used internally by ExportTypedValues; fields are json.RawMessage for flexibility.
type jsonStructValue struct {
	ObjectInfo JSONObjectInfo   `json:"ObjectInfo"`
	Fields     []jsonNamedField `json:"Fields"`
}

// JSONField represents a struct field with name, type, and value.
// Registered with Amino to produce @type tags in JSON output.
// V is json.RawMessage to allow both Value types (amino-serialized) and raw JSON primitives.
type JSONField struct {
	N string          `json:"N"` // Field name
	T Type            `json:"T"` // Type (Amino-serializable)
	V json.RawMessage `json:"V"` // Value as raw JSON (primitives as-is, complex types amino-serialized)
}

// MarshalJSON implements custom JSON marshaling for JSONField to properly embed
// the V field as raw JSON rather than letting amino base64 encode it.
func (jf JSONField) MarshalJSON() ([]byte, error) {
	// Marshal the type using amino for proper @type tags
	tJSON := json.RawMessage("null")
	if jf.T != nil {
		var err error
		tJSON, err = amino.MarshalJSONAny(jf.T)
		if err != nil {
			return nil, err
		}
	}

	// V is already raw JSON, embed it directly
	vJSON := json.RawMessage("null")
	if jf.V != nil {
		vJSON = jf.V
	}

	return json.Marshal(struct {
		N string          `json:"N,omitempty"`
		T json.RawMessage `json:"T"`
		V json.RawMessage `json:"V"`
	}{jf.N, tJSON, vJSON})
}

// JSONObjectInfo is a simplified ObjectInfo for JSON export.
// Avoids verbose nested struct serialization of the full ObjectInfo.
// ID format: ":N" for ephemeral objects, "pkghash:N" for persisted objects.
type JSONObjectInfo struct {
	ID       string `json:"ID"`
	OwnerID  string `json:"OwnerID,omitempty"`
	RefCount int    `json:"RefCount,omitempty"`
}

// makeJSONObjectInfo creates a JSONObjectInfo from ObjectInfo.
// For ephemeral objects (zero ObjectID), uses incrementalID for the ID (":1", ":2", etc).
// For persisted objects, uses the full ObjectID string.
func makeJSONObjectInfo(oi ObjectInfo, incrementalID int) JSONObjectInfo {
	var id string
	if oi.ID.IsZero() && incrementalID > 0 {
		id = fmt.Sprintf(":%d", incrementalID)
	} else {
		id = oi.ID.String() // Uses MarshalAmino format
	}

	var ownerID string
	if !oi.OwnerID.IsZero() {
		ownerID = oi.OwnerID.String()
	}

	return JSONObjectInfo{
		ID:       id,
		OwnerID:  ownerID,
		RefCount: oi.RefCount,
	}
}

// JSONStructValue is a StructValue with named fields for Amino serialization.
// Replaces StructValue during export to include field names in JSON output.
// Implements the Value interface minimally (export-only type).
type JSONStructValue struct {
	ObjectInfo JSONObjectInfo `json:"ObjectInfo"`
	Fields     []JSONField    `json:"Fields"`
}

func (asv *JSONStructValue) assertValue()                     {}
func (asv *JSONStructValue) String() string                   { return "JSONStructValue{...}" }
func (asv *JSONStructValue) DeepFill(store Store) Value       { return asv }
func (asv *JSONStructValue) GetShallowSize() int64            { return 0 }
func (asv *JSONStructValue) VisitAssociated(vis Visitor) bool { return false }

// MarshalJSON implements custom JSON marshaling to properly embed JSONField values.
func (jsv *JSONStructValue) MarshalJSON() ([]byte, error) {
	fields, err := marshalJSONFields(jsv.Fields)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Type       string          `json:"@type"`
		ObjectInfo JSONObjectInfo  `json:"ObjectInfo"`
		Fields     json.RawMessage `json:"Fields"`
	}{
		Type:       "/gno.JSONStructValue",
		ObjectInfo: jsv.ObjectInfo,
		Fields:     fields,
	})
}

// JSONArrayValue is an ArrayValue with human-readable element values for JSON serialization.
// Replaces ArrayValue during export to include proper primitive formatting.
// Implements the Value interface minimally (export-only type).
type JSONArrayValue struct {
	ObjectInfo JSONObjectInfo `json:"ObjectInfo"`
	Elements   []JSONField    `json:"Elements"`
}

func (jav *JSONArrayValue) assertValue()                     {}
func (jav *JSONArrayValue) String() string                   { return "JSONArrayValue{...}" }
func (jav *JSONArrayValue) DeepFill(store Store) Value       { return jav }
func (jav *JSONArrayValue) GetShallowSize() int64            { return 0 }
func (jav *JSONArrayValue) VisitAssociated(vis Visitor) bool { return false }

// MarshalJSON implements custom JSON marshaling to properly embed JSONField values.
func (jav *JSONArrayValue) MarshalJSON() ([]byte, error) {
	elements, err := marshalJSONFields(jav.Elements)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Type       string          `json:"@type"`
		ObjectInfo JSONObjectInfo  `json:"ObjectInfo"`
		Elements   json.RawMessage `json:"Elements"`
	}{
		Type:       "/gno.JSONArrayValue",
		ObjectInfo: jav.ObjectInfo,
		Elements:   elements,
	})
}

// JSONMapValue is a MapValue with human-readable key/value pairs for JSON serialization.
// Replaces MapValue during export to include proper primitive formatting.
// Implements the Value interface minimally (export-only type).
type JSONMapValue struct {
	ObjectInfo JSONObjectInfo `json:"ObjectInfo"`
	Entries    []JSONMapEntry `json:"Entries"`
}

// JSONMapEntry represents a single key-value pair in a JSONMapValue.
type JSONMapEntry struct {
	Key   JSONField `json:"Key"`
	Value JSONField `json:"Value"`
}

func (jmv *JSONMapValue) assertValue()                     {}
func (jmv *JSONMapValue) String() string                   { return "JSONMapValue{...}" }
func (jmv *JSONMapValue) DeepFill(store Store) Value       { return jmv }
func (jmv *JSONMapValue) GetShallowSize() int64            { return 0 }
func (jmv *JSONMapValue) VisitAssociated(vis Visitor) bool { return false }

// MarshalJSON implements custom JSON marshaling to properly embed JSONField values.
func (jmv *JSONMapValue) MarshalJSON() ([]byte, error) {
	// Marshal each entry using struct marshaling
	entriesJSON := make([]json.RawMessage, len(jmv.Entries))
	for i, e := range jmv.Entries {
		keyJSON, err := e.Key.MarshalJSON()
		if err != nil {
			return nil, err
		}
		valJSON, err := e.Value.MarshalJSON()
		if err != nil {
			return nil, err
		}
		entryJSON, err := json.Marshal(struct {
			Key   json.RawMessage `json:"Key"`
			Value json.RawMessage `json:"Value"`
		}{keyJSON, valJSON})
		if err != nil {
			return nil, err
		}
		entriesJSON[i] = entryJSON
	}

	// Combine entries into array
	entriesArrayJSON, _ := json.Marshal(entriesJSON)

	return json.Marshal(struct {
		Type       string          `json:"@type"`
		ObjectInfo JSONObjectInfo  `json:"ObjectInfo"`
		Entries    json.RawMessage `json:"Entries"`
	}{
		Type:       "/gno.JSONMapValue",
		ObjectInfo: jmv.ObjectInfo,
		Entries:    entriesArrayJSON,
	})
}

// primitiveToJSON converts a primitive TypedValue to JSON bytes.
// Returns (json, true) for primitives, (nil, false) for non-primitives.
func primitiveToJSON(tv TypedValue) ([]byte, bool) {
	bt := BaseOf(tv.T)
	pt, ok := bt.(PrimitiveType)
	if !ok {
		return nil, false
	}

	var s string
	switch pt {
	case BoolType, UntypedBoolType:
		s = strconv.FormatBool(tv.GetBool())
	case StringType, UntypedStringType:
		if sv, ok := tv.V.(StringValue); ok {
			s = strconv.Quote(string(sv))
		} else {
			s = strconv.Quote(tv.GetString())
		}
	case IntType:
		s = strconv.FormatInt(int64(tv.GetInt()), 10)
	case Int8Type:
		s = strconv.FormatInt(int64(tv.GetInt8()), 10)
	case Int16Type:
		s = strconv.FormatInt(int64(tv.GetInt16()), 10)
	case Int32Type, UntypedRuneType:
		s = strconv.FormatInt(int64(tv.GetInt32()), 10)
	case Int64Type:
		s = strconv.FormatInt(tv.GetInt64(), 10)
	case UintType:
		s = strconv.FormatUint(uint64(tv.GetUint()), 10)
	case Uint8Type:
		s = strconv.FormatUint(uint64(tv.GetUint8()), 10)
	case DataByteType:
		s = strconv.FormatUint(uint64(tv.GetDataByte()), 10)
	case Uint16Type:
		s = strconv.FormatUint(uint64(tv.GetUint16()), 10)
	case Uint32Type:
		s = strconv.FormatUint(uint64(tv.GetUint32()), 10)
	case Uint64Type:
		s = strconv.FormatUint(tv.GetUint64(), 10)
	case Float32Type:
		s = strconv.FormatFloat(float64(math.Float32frombits(tv.GetFloat32())), 'g', -1, 32)
	case Float64Type:
		s = strconv.FormatFloat(math.Float64frombits(tv.GetFloat64()), 'g', -1, 64)
	case UntypedBigintType:
		if bv, ok := tv.V.(BigintValue); ok {
			s = bv.V.String()
		}
	case UntypedBigdecType:
		if bv, ok := tv.V.(BigdecValue); ok {
			s = bv.V.String()
		}
	}
	if s == "" {
		return nil, false
	}
	return []byte(s), true
}

// marshalJSONFields marshals a slice of JSONField with proper embedding.
func marshalJSONFields(fields []JSONField) ([]byte, error) {
	parts := make([]json.RawMessage, len(fields))
	for i, f := range fields {
		b, err := f.MarshalJSON()
		if err != nil {
			return nil, err
		}
		parts[i] = b
	}
	return json.Marshal(parts)
}

// getElementType extracts element type from array, slice, or pointer types.
func getElementType(typ Type) Type {
	switch t := BaseOf(typ).(type) {
	case *ArrayType:
		return t.Elt
	case *SliceType:
		return t.Elt
	case *PointerType:
		return t.Elt
	}
	return nil
}

// getMapTypes extracts key and value types from a map type.
func getMapTypes(typ Type) (key, val Type) {
	if mt, ok := BaseOf(typ).(*MapType); ok {
		return mt.Key, mt.Value
	}
	return nil, nil
}

// marshalNestedValue marshals a Value that may contain nested JSON types.
// Handles PointerValue, HeapItemValue, and JSON wrapper types.
// The exporter's seen map is needed for HeapItemValue object IDs.
func (e *jsonExporter) marshalNestedValue(v Value) []byte {
	if v == nil {
		return []byte("null")
	}
	switch val := v.(type) {
	case *JSONStructValue:
		if b, err := val.MarshalJSON(); err == nil {
			return b
		}
	case *JSONArrayValue:
		if b, err := val.MarshalJSON(); err == nil {
			return b
		}
	case *JSONMapValue:
		if b, err := val.MarshalJSON(); err == nil {
			return b
		}
	case PointerValue:
		return e.marshalPointerValue(val)
	case *HeapItemValue:
		return e.marshalHeapItemValue(val)
	case *SliceValue:
		return e.marshalSliceValue(val)
	case StringValue:
		// Handle declared string types - output just the string value
		return []byte(strconv.Quote(string(val)))
	case BigintValue:
		// Handle declared bigint types - output the number
		return []byte(val.V.String())
	case BigdecValue:
		// Handle declared bigdec types - output the number
		return []byte(val.V.String())
	}
	return amino.MustMarshalJSONAny(v)
}

// marshalPointerValue marshals a PointerValue to JSON with proper nested type handling.
func (e *jsonExporter) marshalPointerValue(pv PointerValue) []byte {
	baseJSON := e.marshalNestedValue(pv.Base)
	result, _ := json.Marshal(struct {
		Type  string          `json:"@type"`
		TV    *TypedValue     `json:"TV"`
		Base  json.RawMessage `json:"Base"`
		Index string          `json:"Index"`
	}{
		Type:  "/gno.PointerValue",
		TV:    nil,
		Base:  baseJSON,
		Index: strconv.Itoa(pv.Index),
	})
	return result
}

// marshalHeapItemValue marshals a HeapItemValue to JSON with proper nested type handling.
func (e *jsonExporter) marshalHeapItemValue(hiv *HeapItemValue) []byte {
	objID := e.seen[hiv]
	objInfo := makeJSONObjectInfo(hiv.ObjectInfo, objID)

	valueJSON := e.marshalNestedValue(hiv.Value.V)
	typeJSON := amino.MustMarshalJSONAny(hiv.Value.T)

	// Build the inner Value struct
	innerValue, _ := json.Marshal(struct {
		T json.RawMessage `json:"T"`
		V json.RawMessage `json:"V"`
	}{typeJSON, valueJSON})

	result, _ := json.Marshal(struct {
		Type       string          `json:"@type"`
		ObjectInfo JSONObjectInfo  `json:"ObjectInfo"`
		Value      json.RawMessage `json:"Value"`
	}{
		Type:       "/gno.HeapItemValue",
		ObjectInfo: objInfo,
		Value:      innerValue,
	})
	return result
}

// marshalSliceValue marshals a SliceValue to JSON with proper nested type handling.
func (e *jsonExporter) marshalSliceValue(sv *SliceValue) []byte {
	baseJSON := e.marshalNestedValue(sv.Base)
	result, _ := json.Marshal(struct {
		Type   string          `json:"@type"`
		Base   json.RawMessage `json:"Base"`
		Offset string          `json:"Offset"`
		Length string          `json:"Length"`
		Maxcap string          `json:"Maxcap"`
	}{
		Type:   "/gno.SliceValue",
		Base:   baseJSON,
		Offset: strconv.Itoa(sv.Offset),
		Length: strconv.Itoa(sv.Length),
		Maxcap: strconv.Itoa(sv.Maxcap),
	})
	return result
}

// jsonExporter handles JSON serialization of Gno values.
// Configuration fields control export behavior.
type jsonExporter struct {
	opts JSONExporterOptions

	// state (set during export)
	store Store
	seen  map[Object]int

	// structTypes maps StructValue pointers to their corresponding StructType
	// This allows preserving field names during JSON serialization
	structTypes map[*StructValue]*StructType
}

// init initializes internal state for a new export operation.
func (e *jsonExporter) init() {
	if e.seen == nil {
		e.seen = map[Object]int{}
	}
	if e.structTypes == nil {
		e.structTypes = map[*StructValue]*StructType{}
	}
}

// jsonExportedTypedValueFromExported converts an already-exported TypedValue to JSON format.
// This is used by JSONExportTypedValues() after calling jsonExportTypedValue() which
// expands ephemeral objects inline while keeping RefValue for persisted objects.
func (e *jsonExporter) jsonExportedTypedValueFromExported(tv TypedValue) *jsonTypedValue {
	return &jsonTypedValue{
		Type:  jsonExportedType(tv.T),
		Value: e.jsonExportedValueFromExported(tv),
	}
}

// jsonExportedValueFromExported converts an already-exported TypedValue's value to JSON.
// Unlike jsonExportedValue(), this handles TypedValues that have already been processed
// by jsonExportTypedValue() - meaning primitive N values have NOT been extracted yet.
func (e *jsonExporter) jsonExportedValueFromExported(tv TypedValue) []byte {
	// Handle primitives using shared helper
	if b, ok := primitiveToJSON(tv); ok {
		return b
	}

	if tv.V == nil {
		return []byte("null")
	}

	// Special handling for StructValue to include field names
	if sv, ok := tv.V.(*StructValue); ok {
		return e.marshalStructValueWithNames(sv)
	}

	// Use shared nested value marshaler for JSON types and pointers
	return e.marshalNestedValue(tv.V)
}

// marshalStructValueWithNames marshals a StructValue to JSON with field names included.
// Uses the exporter's structTypes mapping to get field name information.
func (e *jsonExporter) marshalStructValueWithNames(sv *StructValue) []byte {
	// Look up the struct type from the mapping
	structType := e.structTypes[sv]

	// Build named fields
	namedFields := make([]jsonNamedField, 0, len(sv.Fields))
	for i, ftv := range sv.Fields {
		var fieldName string
		if structType != nil && i < len(structType.Fields) {
			fieldName = string(structType.Fields[i].Name)
			// Use json tag name if available
			if jsonName := getJSONFieldName(structType.Fields[i].Tag); jsonName != "" {
				fieldName = jsonName
			}
		} else {
			fieldName = fmt.Sprintf("field%d", i)
		}

		namedFields = append(namedFields, jsonNamedField{
			Name:  fieldName,
			Type:  jsonExportedType(ftv.T),
			Value: e.jsonExportedValueFromExported(ftv),
		})
	}

	// Get incremental ID from seen map for ephemeral objects
	objID := e.seen[sv]

	jsv := jsonStructValue{
		ObjectInfo: makeJSONObjectInfo(sv.ObjectInfo, objID),
		Fields:     namedFields,
	}

	result, err := json.Marshal(jsv)
	if err != nil {
		return amino.MustMarshalJSONAny(sv) // fallback
	}
	return result
}

// getJSONFieldName extracts the field name from a json struct tag.
// For example, `json:"myName,omitempty"` returns "myName".
// Returns empty string if no valid json tag is found.
func getJSONFieldName(tag Tag) string {
	jsonTag := reflect.StructTag(tag).Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}
	// Extract name before comma (for options like `json:"name,omitempty"`)
	if commaIdx := strings.Index(jsonTag, ","); commaIdx >= 0 {
		jsonTag = jsonTag[:commaIdx]
	}
	if jsonTag == "-" {
		return ""
	}
	return jsonTag
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
	jsonTag := reflect.StructTag(tag).Get("json")
	return jsonTag == "-" || strings.HasPrefix(jsonTag, "-,")
}

// getStructTypeFromType resolves a Type to its underlying *StructType.
// Returns nil if the type is not a struct type.
// Note: This function cannot resolve RefType - use jsonExporter.resolveStructType() instead.
func getStructTypeFromType(typ Type) *StructType {
	if typ == nil {
		return nil
	}

	switch t := typ.(type) {
	case *DeclaredType:
		return getStructTypeFromType(t.Base)
	case *StructType:
		return t
	case *PointerType:
		return getStructTypeFromType(t.Elt)
	case RefType:
		// RefType needs to be resolved via store, which we don't have here.
		// Caller should use jsonExporter.resolveStructType() instead.
		return nil
	default:
		return nil
	}
}

// resolveStructType resolves a Type to its underlying *StructType.
// Unlike getStructTypeFromType, this method can resolve RefType using the store.
// RefType is used for declared types (e.g., gno.land/r/demo/json_export.SimpleStruct)
// and needs store access to retrieve the actual type definition.
func (e *jsonExporter) resolveStructType(typ Type) *StructType {
	if typ == nil {
		return nil
	}

	// If it's a RefType, try to resolve it via store first
	if rt, ok := typ.(RefType); ok {
		if e.store != nil {
			if resolvedType := e.store.GetType(rt.ID); resolvedType != nil {
				return getStructTypeFromType(resolvedType)
			}
		}
		return nil
	}

	return getStructTypeFromType(typ)
}

// ExportObject exports an Object to JSON using Amino marshaling.
// This function handles all object types including HeapItemValue, StructValue,
// ArrayValue, MapValue.
//
// The export process:
//  1. Root object is always fully expanded
//  2. HeapItemValue is automatically unwrapped to show the underlying value
//     (HeapItemValue is an implementation detail for pointer indirection)
//  3. Nested real objects (persisted) become RefValue with their actual ObjectID
//  4. Nested non-real objects are expanded inline (no queryable ObjectID)
//  5. Cycles are detected and converted to RefValue to prevent infinite recursion
//
// The exported object is serialized via Amino's MarshalJSONAny which
// produces @type tags like "/gno.StructValue" with full ObjectInfo.
func (e *jsonExporter) ExportObject(m *Machine, obj Object) ([]byte, error) {
	e.init()

	if m != nil {
		e.store = m.Store
	}

	// Extract type and ObjectInfo from HeapItemValue BEFORE unwrapping.
	// HeapItemValue.Value is a TypedValue containing both type (.T) and value (.V).
	// The type is needed for field name extraction in struct exports.
	// The ObjectInfo is preserved to show the wrapper's position in the ownership tree,
	// not the inner value's ObjectInfo (which just points back to the wrapper).
	var typ Type
	var wrapperObjInfo *ObjectInfo
	if hiv, ok := obj.(*HeapItemValue); ok {
		typ = hiv.Value.T
		wrapperObjInfo = hiv.GetObjectInfo()
	}

	// Unwrap HeapItemValue - it's an implementation detail that users shouldn't see.
	// When querying a HeapItemValue, show the underlying value instead.
	obj = unwrapHeapItemValue(e.store, obj)

	// If we still don't have type info and the object has an owner,
	// try to find the type by looking at the owner's fields.
	if typ == nil && e.store != nil {
		typ = e.findTypeFromOwner(obj)
	}

	// Export the object using Amino-specific export that:
	// - Expands root and non-real objects inline
	// - Converts real nested objects to RefValue with their actual ObjectID
	// - Converts StructValue to JSONStructValue with field names
	// The typ parameter provides type info for field name extraction.
	exported := e.exportObjectValue(obj, typ, 0)

	// If we unwrapped a HeapItemValue, use its ObjectInfo instead of the inner value's.
	// This ensures the ObjectInfo shows the wrapper's position in the ownership tree.
	if wrapperObjInfo != nil {
		exported = replaceObjectInfo(exported, *wrapperObjInfo)
	}

	// Use custom MarshalJSON for our JSON types to avoid Amino base64 encoding of json.RawMessage
	switch v := exported.(type) {
	case *JSONStructValue:
		return v.MarshalJSON()
	case *JSONArrayValue:
		return v.MarshalJSON()
	case *JSONMapValue:
		return v.MarshalJSON()
	default:
		// Fallback to Amino for other types
		return amino.MarshalJSONAny(exported)
	}
}

// findTypeFromOwner attempts to find the type of an object by looking at its owner.
// When an object is stored as a field in a struct, the type info is in the parent.
// This function loads the owner and searches for a field that references this object.
func (e *jsonExporter) findTypeFromOwner(obj Object) Type {
	// Use ObjectInfo.OwnerID directly, not GetOwnerID() which requires runtime owner pointer
	ownerID := obj.GetObjectInfo().OwnerID
	if ownerID.IsZero() {
		return nil
	}

	owner := e.store.GetObjectSafe(ownerID)
	if owner == nil {
		return nil
	}

	objID := obj.GetObjectID()

	// Check if owner is a StructValue with fields referencing this object
	if sv, ok := owner.(*StructValue); ok {
		// We need the owner's type to know the field types
		ownerType := e.findTypeFromOwner(owner)
		ownerStructType := e.resolveStructType(ownerType)

		for i, ftv := range sv.Fields {
			// Check if this field's value references our object
			if rv, ok := ftv.V.(RefValue); ok && rv.ObjectID == objID {
				if ownerStructType != nil && i < len(ownerStructType.Fields) {
					return ownerStructType.Fields[i].Type
				}
			}
			// Also check if this field IS our object (inline struct)
			if fieldObj, ok := ftv.V.(Object); ok && fieldObj.GetObjectID() == objID {
				if ownerStructType != nil && i < len(ownerStructType.Fields) {
					return ownerStructType.Fields[i].Type
				}
			}
		}
	}

	// Check if owner is a HeapItemValue (pointer target)
	if hiv, ok := owner.(*HeapItemValue); ok {
		// The HeapItemValue stores the type in Value.T
		if rv, ok := hiv.Value.V.(RefValue); ok {
			if rv.ObjectID == objID {
				return hiv.Value.T
			}
		}
		if innerObj, ok := hiv.Value.V.(Object); ok {
			if innerObj.GetObjectID() == objID {
				return hiv.Value.T
			}
		}
	}

	// Check if owner is an ArrayValue
	if av, ok := owner.(*ArrayValue); ok {
		for _, etv := range av.List {
			if rv, ok := etv.V.(RefValue); ok && rv.ObjectID == objID {
				return etv.T
			}
			if elemObj, ok := etv.V.(Object); ok && elemObj.GetObjectID() == objID {
				return etv.T
			}
		}
	}

	return nil
}

// JSONExportObject exports an Object using default options.
// For custom options, use JSONExporterOptions.ExportObject() directly.
func JSONExportObject(m *Machine, obj Object) ([]byte, error) {
	return (&jsonExporter{}).ExportObject(m, obj)
}

// ExportObject exports an Object using the configured options.
func (opts JSONExporterOptions) ExportObject(m *Machine, obj Object) ([]byte, error) {
	e := &jsonExporter{opts: opts}
	return e.ExportObject(m, obj)
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

// replaceObjectInfo replaces the ObjectInfo in an exported Value with the provided ObjectInfo.
// This is used when unwrapping HeapItemValue to preserve the wrapper's position in the ownership tree.
func replaceObjectInfo(v Value, oi ObjectInfo) Value {
	jsonOI := makeJSONObjectInfo(oi, 0)
	switch val := v.(type) {
	case *JSONStructValue:
		val.ObjectInfo = jsonOI
		return val
	case *JSONArrayValue:
		val.ObjectInfo = jsonOI
		return val
	case *JSONMapValue:
		val.ObjectInfo = jsonOI
		return val
	}
	return v
}

// exportObjectValue exports an Object for JSON serialization.
// Root object (depth 0) is always expanded inline.
// Nested real objects (depth > 0) become RefValue with queryable ObjectID.
// Non-real objects are always expanded inline.
func (e *jsonExporter) exportObjectValue(obj Object, typ Type, depth int) Value {
	if obj == nil {
		return nil
	}

	// Unwrap HeapItemValue - it's an implementation detail.
	// For nested real HeapItemValue, return RefValue.
	if hiv, ok := obj.(*HeapItemValue); ok {
		if depth > 0 && hiv.GetIsReal() {
			return RefValue{
				ObjectID: hiv.GetObjectID(),
				Escaped:  hiv.GetIsEscaped() || hiv.GetIsNewEscaped(),
				Hash:     hiv.GetHash(),
			}
		}
		// Unwrap for root or non-real
		obj = unwrapHeapItemValue(e.store, obj)
	}

	// Check for cycles first
	if id, exists := e.seen[obj]; exists {
		// Cycle detected - return RefValue
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

	// For nested real objects, return RefValue with queryable ObjectID.
	if depth > 0 && obj.GetIsReal() {
		return RefValue{
			ObjectID: obj.GetObjectID(),
			Escaped:  obj.GetIsEscaped() || obj.GetIsNewEscaped(),
			Hash:     obj.GetHash(),
		}
	}

	// Check depth limit (MaxDepth = 0 means no limit)
	if e.opts.MaxDepth > 0 && depth > e.opts.MaxDepth {
		if obj.GetIsReal() {
			return RefValue{
				ObjectID: obj.GetObjectID(),
				Escaped:  true,
			}
		}
		// Non-real object at depth limit - assign local ID
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

	// Expand the object
	return e.exportCopyValue(obj, typ, depth, id)
}

// exportCopyValue creates an exported copy of a Value for Amino JSON serialization.
// The typ parameter provides type information for field name extraction.
// The objID parameter is the incremental ID assigned to ephemeral objects.
// For StructValue, converts to JSONStructValue with field names when type info is available.
func (e *jsonExporter) exportCopyValue(val Value, typ Type, depth int, objID int) Value {
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
		eltType := getElementType(typ)
		if cv.Data == nil {
			// Convert to JSONArrayValue with human-readable elements
			elements := make([]JSONField, len(cv.List))
			for i, etv := range cv.List {
				// Only override type if we have element type info
				if eltType != nil {
					etv.T = eltType
				}
				exported := e.exportTypedValue(etv, depth+1)
				elements[i] = JSONField{
					N: fmt.Sprintf("%d", i),
					T: exported.T,
					V: e.extractPrimitiveOrValueJSON(exported),
				}
			}
			return &JSONArrayValue{
				ObjectInfo: makeJSONObjectInfo(cv.ObjectInfo, objID),
				Elements:   elements,
			}
		}
		return &ArrayValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Data:       cp(cv.Data),
		}
	case *SliceValue:
		eltType := getElementType(typ)
		var base Value
		if cv.Base != nil {
			if obj, ok := cv.Base.(Object); ok {
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
		// Get struct type for field names (resolves RefType via store if needed)
		structType := e.resolveStructType(typ)

		// Build JSONStructValue with field names
		namedFields := make([]JSONField, 0, len(cv.Fields))
		for i, ftv := range cv.Fields {
			var fieldName string
			var fieldType Type

			if structType != nil && i < len(structType.Fields) {
				ft := structType.Fields[i]

				// Skip json:"-" tagged fields
				if hasJSONSkipTag(ft.Tag) {
					continue
				}

				// Check export visibility
				if !e.opts.ExportUnexported && !isExportedName(ft.Name) {
					continue
				}

				fieldName = string(ft.Name)
				// Use json tag name if available
				if jsonName := getJSONFieldName(ft.Tag); jsonName != "" {
					fieldName = jsonName
				}
				fieldType = ft.Type
			} else {
				fieldName = fmt.Sprintf("field%d", i)
				fieldType = ftv.T
			}

			// Export the field value
			ftv.T = fieldType
			exportedField := e.exportTypedValue(ftv, depth+1)

			namedFields = append(namedFields, JSONField{
				N: fieldName,
				T: exportedField.T,
				V: e.extractPrimitiveOrValueJSON(exportedField),
			})
		}

		return &JSONStructValue{
			ObjectInfo: makeJSONObjectInfo(cv.ObjectInfo, objID),
			Fields:     namedFields,
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
		fnc := e.exportCopyValue(cv.Func, nil, depth, 0).(*FuncValue)
		rtv := e.exportTypedValue(cv.Receiver, depth+1)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		keyType, valType := getMapTypes(typ)
		// Convert to JSONMapValue with human-readable entries
		entries := make([]JSONMapEntry, 0)
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			k := cur.Key
			v := cur.Value
			// Only override types if we have type info
			if keyType != nil {
				k.T = keyType
			}
			if valType != nil {
				v.T = valType
			}
			keyExported := e.exportTypedValue(k, depth+1)
			valExported := e.exportTypedValue(v, depth+1)
			entries = append(entries, JSONMapEntry{
				Key: JSONField{
					T: keyExported.T,
					V: e.extractPrimitiveOrValueJSON(keyExported),
				},
				Value: JSONField{
					T: valExported.T,
					V: e.extractPrimitiveOrValueJSON(valExported),
				},
			})
		}
		return &JSONMapValue{
			ObjectInfo: makeJSONObjectInfo(cv.ObjectInfo, objID),
			Entries:    entries,
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

// exportTypedValue exports a TypedValue for Amino JSON serialization.
func (e *jsonExporter) exportTypedValue(tv TypedValue, depth int) TypedValue {
	result := TypedValue{}
	if tv.T != nil {
		result.T = exportRefOrCopyType(tv.T, e.seen)
	}

	if obj, ok := tv.V.(Object); ok {
		result.V = e.exportObjectValue(obj, tv.T, depth)
		return result
	}

	// For non-Object values, use the Amino export copy
	if tv.V != nil {
		result.V = e.exportCopyValue(tv.V, tv.T, depth, 0)
	}
	// Copy primitive values
	result.N = tv.N
	return result
}

// extractPrimitiveOrValueJSON extracts human-readable values from primitives as json.RawMessage,
// or returns the amino-serialized Value for complex types.
// Used by JSONArrayValue, JSONMapValue, and JSONStructValue to ensure consistent primitive formatting.
func (e *jsonExporter) extractPrimitiveOrValueJSON(tv TypedValue) json.RawMessage {
	// Handle primitives using shared helper
	if b, ok := primitiveToJSON(tv); ok {
		return json.RawMessage(b)
	}
	// For non-primitives, use shared nested value marshaler
	if tv.V == nil {
		return json.RawMessage("null")
	}
	return json.RawMessage(e.marshalNestedValue(tv.V))
}

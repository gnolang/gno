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
	// 0 mean no limit.
	MaxDepth int
}

// ExportTypedValues exports multiple TypedValues to JSON.
// Persisted objects are shown as RefValue (with queryable ObjectID).
// Ephemeral objects are expanded inline.
func (opts JSONExporterOptions) ExportTypedValues(tvs []TypedValue) ([]byte, error) {
	e := &jsonExporter{opts: opts}
	e.init()

	jexps := make([]*jsonTypedValue, len(tvs))

	for i, tv := range tvs {
		// Use JSON-specific export that:
		// - Shows RefValue for persisted (real) objects (with queryable ObjectID)
		// - Expands ephemeral (unreal) objects inline (no ObjectID)
		// - Filters unexported fields and json:"-" tagged fields based on options
		exported := e.exportTypedValue(tv, 0)
		jexps[i] = e.jsonExportedTypedValueFromExported(exported)
	}

	return json.Marshal(jexps)
}

// JSONExportTypedValues exports TypedValues using default options.
// For custom options, use JSONExporter directly.
func JSONExportTypedValues(tvs []TypedValue) ([]byte, error) {
	return JSONExporterOptions{}.ExportTypedValues(tvs)
}

func JSONExportTypedValue(tv TypedValue, seen map[Object]int) ([]byte, error) {
	if seen == nil {
		seen = map[Object]int{}
	}

	tv = exportValue(tv, seen) // first export value
	return json.Marshal(jsonExportedTypedValue(tv))
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

// jsonNamedField represents a struct field with its name for JSON export.
type jsonNamedField struct {
	Name  string          `json:"N"`
	Type  json.RawMessage `json:"T"`
	Value json.RawMessage `json:"V"`
}

// jsonStructValue is a custom JSON representation of StructValue with field names.
// Uses AminoObjectInfo for cleaner output of ephemeral objects.
type jsonStructValue struct {
	ObjectInfo AminoObjectInfo  `json:"ObjectInfo"`
	Fields     []jsonNamedField `json:"Fields"`
}

// AminoNamedField is a TypedValue with field name for Amino serialization.
// This type is registered with Amino to produce proper @type tags.
type AminoNamedField struct {
	N string `json:"N"` // Field name
	T Type   `json:"T"` // Type (Amino-serializable)
	V Value  `json:"V"` // Value (Amino-serializable)
}

// AminoObjectInfo is a simplified ObjectInfo for JSON export.
// For ephemeral objects: ID = ":N" (incremental, e.g., ":1", ":2")
// For real objects: ID = "pkghash:N" (full ObjectID string representation)
// This avoids the verbose nested struct serialization of the full ObjectInfo.
type AminoObjectInfo struct {
	ID       string `json:"ID"`
	RefCount int    `json:"RefCount,omitempty"`
}

// makeAminoObjectInfo creates an AminoObjectInfo with proper ID formatting.
// For ephemeral objects (zero ObjectID), uses the incrementalID: ":1", ":2", etc.
// For real objects, uses the full ObjectID string representation.
func makeAminoObjectInfo(oi ObjectInfo, incrementalID int) AminoObjectInfo {
	var id string
	if oi.ID.IsZero() && incrementalID > 0 {
		id = fmt.Sprintf(":%d", incrementalID)
	} else {
		id = oi.ID.String() // Uses MarshalAmino format
	}
	return AminoObjectInfo{
		ID:       id,
		RefCount: oi.RefCount,
	}
}

// AminoStructValue is a StructValue with named fields for Amino serialization.
// This type replaces StructValue during export to include field names.
// It implements the Value interface minimally since it's only used for JSON export.
type AminoStructValue struct {
	ObjectInfo AminoObjectInfo   `json:"ObjectInfo"`
	Fields     []AminoNamedField `json:"Fields"`
}

// Value interface implementation for AminoStructValue (export-only type).
func (asv *AminoStructValue) assertValue()                      {}
func (asv *AminoStructValue) String() string                    { return "AminoStructValue{...}" }
func (asv *AminoStructValue) DeepFill(store Store) Value        { return asv }
func (asv *AminoStructValue) GetShallowSize() int64             { return 0 }
func (asv *AminoStructValue) VisitAssociated(vis Visitor) bool  { return false }

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
	if e.opts.MaxDepth == 0 {
		e.opts.MaxDepth = DefaultMaxDepth
	}
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

		// Special handling for StructValue to include field names
		if sv, ok := tv.V.(*StructValue); ok {
			return e.marshalStructValueWithNames(sv)
		}

		return amino.MustMarshalJSONAny(tv.V)
	}
}

// marshalStructValueWithNames marshals a StructValue to JSON with field names included.
// Uses the exporter's structTypes mapping to get field name information.
func (e *jsonExporter) marshalStructValueWithNames(sv *StructValue) []byte {
	// Look up the struct type from the mapping
	structType := e.structTypes[sv]

	// Build named fields
	var namedFields []jsonNamedField
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
		ObjectInfo: makeAminoObjectInfo(sv.ObjectInfo, objID),
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
	if tag == "" {
		return ""
	}
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

		// Scan to colon
		i = 0
		for i < len(tagStr) && tagStr[i] != ':' && tagStr[i] != ' ' {
			i++
		}
		if i >= len(tagStr) {
			break
		}
		key := tagStr[:i]
		tagStr = tagStr[i:]

		if len(tagStr) < 2 || tagStr[0] != ':' || tagStr[1] != '"' {
			break
		}
		tagStr = tagStr[2:] // skip :"

		// Find closing quote
		i = 0
		for i < len(tagStr) && tagStr[i] != '"' {
			if tagStr[i] == '\\' && i+1 < len(tagStr) {
				i++ // skip escaped char
			}
			i++
		}
		if i >= len(tagStr) {
			break
		}
		value := tagStr[:i]
		tagStr = tagStr[i+1:]

		if key == "json" {
			// Extract name before comma
			if commaIdx := strings.Index(value, ","); commaIdx >= 0 {
				value = value[:commaIdx]
			}
			if value != "" && value != "-" {
				return value
			}
			return ""
		}
	}
	return ""
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

	// Extract type from HeapItemValue BEFORE unwrapping.
	// HeapItemValue.Value is a TypedValue containing both type (.T) and value (.V).
	// The type is needed for field name extraction in struct exports.
	var typ Type
	if hiv, ok := obj.(*HeapItemValue); ok {
		typ = hiv.Value.T
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
	// - Converts StructValue to AminoStructValue with field names
	// The typ parameter provides type info for field name extraction.
	exported := e.aminoExportObjectValue(obj, typ, 0)

	// Now serialize the exported (cycle-safe) object via Amino.
	return amino.MarshalJSONAny(exported)
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

// aminoExportObjectValue exports an Object for JSON serialization via Amino.
// Unlike exportCopyValueWithRefs (for persistence), this function:
// - Always expands the root object (depth 0)
// - Unwraps HeapItemValue to show the underlying object directly
// - Expands non-real objects inline (they can't be queried separately)
// - Converts real nested objects to RefValue with actual ObjectID (queryable)
// - Converts StructValue to AminoStructValue with field names when type info is available
func (e *jsonExporter) aminoExportObjectValue(obj Object, typ Type, depth int) Value {
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
		obj = unwrapHeapItemValue(e.store, obj)
	}

	// Check for cycles first
	if id, exists := e.seen[obj]; exists {
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
	if e.opts.MaxDepth >= 0 && depth > e.opts.MaxDepth {
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

	// Expand the object, passing the incremental ID for ephemeral object formatting
	return e.aminoExportCopyValue(obj, typ, depth, id)
}

// aminoExportCopyValue creates an exported copy of a Value for Amino JSON serialization.
// The typ parameter provides type information for field name extraction.
// The objID parameter is the incremental ID assigned to ephemeral objects.
// For StructValue, converts to AminoStructValue with field names when type info is available.
func (e *jsonExporter) aminoExportCopyValue(val Value, typ Type, depth int, objID int) Value {
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
			base = e.aminoExportObjectValue(b, eltType, depth+1)
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
				etv.T = eltType // Set element type
				list[i] = e.aminoExportTypedValue(etv, depth+1)
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
				var arrType Type
				if eltType != nil {
					arrType = &ArrayType{Elt: eltType}
				}
				base = e.aminoExportObjectValue(obj, arrType, depth+1)
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

		// Build AminoStructValue with field names
		namedFields := make([]AminoNamedField, 0, len(cv.Fields))
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
			exportedField := e.aminoExportTypedValue(ftv, depth+1)

			namedFields = append(namedFields, AminoNamedField{
				N: fieldName,
				T: exportedField.T,
				V: exportedField.V,
			})
		}

		return &AminoStructValue{
			ObjectInfo: makeAminoObjectInfo(cv.ObjectInfo, objID),
			Fields:     namedFields,
		}
	case *FuncValue:
		source := toRefNode(cv.Source)
		var parent Value
		if cv.Parent != nil {
			if obj, ok := cv.Parent.(Object); ok {
				parent = e.aminoExportObjectValue(obj, nil, depth+1)
			} else if ref, ok := cv.Parent.(RefValue); ok {
				parent = ref
			}
		}
		captures := make([]TypedValue, len(cv.Captures))
		for i, ctv := range cv.Captures {
			captures[i] = e.aminoExportTypedValue(ctv, depth+1)
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
		fnc := e.aminoExportCopyValue(cv.Func, nil, depth, 0).(*FuncValue)
		rtv := e.aminoExportTypedValue(cv.Receiver, depth+1)
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
		}
	case *MapValue:
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			key2 := e.aminoExportTypedValue(cur.Key, depth+1)
			val2 := e.aminoExportTypedValue(cur.Value, depth+1)
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
			vals[i] = e.aminoExportTypedValue(tv, depth+1)
		}
		var bparent Value
		if cv.Parent != nil {
			if obj, ok := cv.Parent.(Object); ok {
				bparent = e.aminoExportObjectValue(obj, nil, depth+1)
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
			Value:      e.aminoExportTypedValue(cv.Value, depth+1),
		}
	default:
		panic(fmt.Sprintf("unexpected type %v", reflect.TypeOf(val)))
	}
}

// aminoExportTypedValue exports a TypedValue for Amino JSON serialization.
func (e *jsonExporter) aminoExportTypedValue(tv TypedValue, depth int) TypedValue {
	result := TypedValue{}
	if tv.T != nil {
		result.T = exportRefOrCopyType(tv.T, e.seen)
	}

	if obj, ok := tv.V.(Object); ok {
		result.V = e.aminoExportObjectValue(obj, tv.T, depth)
		return result
	}

	// For non-Object values, use the Amino export copy
	if tv.V != nil {
		result.V = e.aminoExportCopyValue(tv.V, tv.T, depth, 0)
	}
	// Copy primitive values
	result.N = tv.N
	return result
}

// exportTypedValue exports a TypedValue for JSON serialization.
// Returns RefValue for persisted (real) objects at all levels.
// Only ephemeral (unreal) objects are expanded inline.
func (e *jsonExporter) exportTypedValue(tv TypedValue, depth int) TypedValue {
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
func (e *jsonExporter) exportObjectValue(obj Object, typ Type, depth int) Value {
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
	if e.opts.MaxDepth >= 0 && depth > e.opts.MaxDepth {
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
func (e *jsonExporter) exportCopyValue(val Value, typ Type, depth int) Value {
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
		// Get struct type for field filtering (resolves RefType via store if needed)
		structType := e.resolveStructType(typ)

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
		var filteredStructType StructType // Build a filtered struct type with only included fields
		for i, ftv := range cv.Fields {
			ft := structType.Fields[i]

			// Always skip json:"-" tagged fields
			if hasJSONSkipTag(ft.Tag) {
				continue
			}

			// Check export visibility
			if !e.opts.ExportUnexported && !isExportedName(ft.Name) {
				continue
			}

			// Export this field with its type
			ftv.T = ft.Type
			fields = append(fields, e.exportTypedValue(ftv, depth+1))
			filteredStructType.Fields = append(filteredStructType.Fields, ft)
		}
		result := &StructValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Fields:     fields,
		}
		// Store the struct type mapping for later JSON serialization
		e.structTypes[result] = &filteredStructType
		return result
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

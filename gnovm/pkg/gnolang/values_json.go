package gnolang

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/encoding/json"
)

const defaultIndent = "  "
const defaultDepthLimit = 10000

func (tv *TypedValue) MarshalJSON() ([]byte, error) {
	return MarshalOptions{}.Marshal(tv)
}

func (tv *TypedValue) MarshalJSONAmino() ([]byte, error) {
	return MarshalOptions{AminoFormat: true}.Marshal(tv)
}

func (tv *TypedValue) UnmarshalJSON(b []byte) error {
	return UnmarshalOptions{}.Unmarshal(b, tv)
}

func (tv *TypedValue) UnmarshalJSONAmino(b []byte) error {
	return UnmarshalOptions{AminoFormat: true}.Unmarshal(b, tv)
}

// MarshalOptions is a configurable JSON format marshaler.
type MarshalOptions struct {
	// Format an output compatible with amino
	AminoFormat bool

	// Multiline specifies whether the marshaler should format the output in
	// indented-form with every textual element on a new line.
	// If Indent is an empty string, then an arbitrary indent is chosen.
	Multiline bool

	// Indent specifies the set of indentation characters to use in a multiline
	// formatted output such that every entry is preceded by Indent and
	// terminated by a newline. If non-empty, then Multiline is treated as true.
	// Indent can only be composed of space or tab characters.
	Indent string

	// If this is true Object will be Wnwraped when encounter, instead of
	// producing an {@type: <str>, "oid": <str>} obj
	FillRefValue bool

	// XXX: TODO
	Store Store

	// XXX: TODO
	Alloc *Allocator
}

// UnmarshalOptions is a configurable JSON format parser.
type UnmarshalOptions struct {
	// Format an output compatible with amino
	AminoFormat bool

	// If AllowPartial is set, input for messages that will result in missing
	// required fields will not return an error.
	AllowPartial bool

	// If DiscardUnknown is set, unknown fields and enum name values are ignored.
	DiscardUnknown bool

	// DepthLimit limits how deeply messages may be nested.
	// If zero, a default limit is applied.
	DepthLimit int

	// XXX: TODO
	Store Store

	// XXX: TODO
	Alloc *Allocator
}

func (o MarshalOptions) Marshal(tv *TypedValue) ([]byte, error) {
	return o.marshal(nil, tv)
}

// Unmarshal reads the given []byte and populates the given [TypedValue]
// using options in the UnmarshalOptions object.
// It will clear the Value first.
// For now Type T must be set to be able to unarshal value from byte.
func (o UnmarshalOptions) Unmarshal(b []byte, tv *TypedValue) error {
	return o.unmarshal(b, tv)
}

// unmarshal is a centralized function that all unmarshal operations go through.
func (o UnmarshalOptions) unmarshal(b []byte, tv *TypedValue) error {
	// tv.Reset()  XXX: reset typed value ?
	if o.Alloc == nil {
		o.Alloc = nilAllocator
	}

	if o.Store == nil {
		o.Store = NewStore(o.Alloc, nil, nil)
	}

	if o.DepthLimit == 0 {
		o.DepthLimit = defaultDepthLimit
	}

	dec := decoder{json.NewDecoder(b), o}
	if err := dec.unmarshalValue(tv); err != nil {
		return err
	}

	// Check for EOF.
	tok, err := dec.Read()
	if err != nil {
		return err
	}
	if tok.Kind() != json.EOF {
		return dec.unexpectedTokenError(tok)
	}

	return nil
}

// MarshalAppend appends the JSON format encoding of m to b,
// returning the result.
func (o MarshalOptions) MarshalAppend(b []byte, m *TypedValue) ([]byte, error) {
	return o.marshal(b, m)
}

// marshal is a centralized function that all marshal operations go through.
func (o MarshalOptions) marshal(b []byte, tv *TypedValue) ([]byte, error) {
	if o.Multiline && o.Indent == "" {
		o.Indent = defaultIndent
	}

	if o.Alloc == nil {
		o.Alloc = nilAllocator
	}

	if o.Store == nil {
		o.Store = NewStore(o.Alloc, nil, nil)
	}

	internalEnc, err := json.NewEncoder(b, o.Indent)
	if err != nil {
		return nil, err
	}

	// Treat nil message interface as an empty message,
	// in which case the output in an empty JSON object.
	if tv == nil {
		return append(b, '{', '}'), nil
	}

	enc := encoder{internalEnc, map[string]bool{}, o}
	if err := enc.marshalValue(tv); err != nil {
		return nil, err
	}

	return internalEnc.Bytes(), nil
}

type encoder struct {
	*json.Encoder
	cache map[string]bool
	opts  MarshalOptions
}

func (e encoder) store() Store {
	return e.opts.Store
}

type decoder struct {
	*json.Decoder
	opts UnmarshalOptions
}

// newError returns an error object with position info.
func (d decoder) newError(pos int, f string, x ...any) error {
	line, column := d.Position(pos)
	head := fmt.Sprintf("(line %d:%d): ", line, column)
	return fmt.Errorf(head+f, x...)
}

// unexpectedTokenError returns a syntax error for the given unexpected token.
func (d decoder) unexpectedTokenError(tok json.Token) error {
	return d.syntaxError(tok.Pos(), "unexpected token %s", tok.RawString())
}

// syntaxError returns a syntax error for given position.
func (d decoder) syntaxError(pos int, f string, x ...any) error {
	line, column := d.Position(pos)
	head := fmt.Sprintf("syntax error (line %d:%d): ", line, column)
	return fmt.Errorf(head+f, x...)
}

// marshalValue marshals the fields in the given TypedValue.
func (e encoder) marshalValue(tv *TypedValue) error {
	if tv.T == nil {
		e.WriteNull()
		return nil
	}

	switch tv.T.Kind() {
	case BoolKind, StringKind,
		IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind,
		UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind,
		Float32Kind, Float64Kind,
		BigintKind, BigdecKind:
		return e.marshalScalar(tv)

	case StructKind:
		return e.marshalStructValue(tv)

	case ArrayKind, SliceKind, TupleKind: // List
		return e.marshalListValue(tv)

	case InterfaceKind:
		return e.marshalInterfaceValue(tv)

	case PointerKind:
		return e.marshalPointerValue(tv)
	default:
		return fmt.Errorf("unable to marshal unknown type: %q", tv.T.Kind())
	}

	// fmt.Printf("%+#v\n", tv.V)
	// // store := e.opts.Store

	// var typeURL string
	// var oid string
	// // switch cv := tv.V.(type) {
	// // case TypeValue:

	// }

	// v := copyValueWithRefs(tv.V)
	// fmt.Printf("%#v\n", v)

	// if ctv, ok := tv.V.(TypeValue); ok {
	// 	switch cv := ctv.Type.(type) {
	// 	case *DeclaredType:

	// 		fmt.Println(cv)
	// 	}
	// }

	// default:
	// 	panic("NOOO")

	// case RefValue:
	// 	// XXX: check for empty pkgpath
	// 	typeURL = cv.PkgPath
	// 	oid = cv.ObjectID.String()
	// case PointerValue:
	// 	if ref, ok := cv.Base.(RefValue); ok {
	// 		base := store.GetObject(ref.ObjectID).(Value)

	// 		cv.Base = base
	// 		switch cb := base.(type) {
	// 		case *ArrayValue:
	// 			et := baseOf(tv.T).(*PointerType).Elt
	// 			epv := cb.GetPointerAtIndexInt2(store, cv.Index, et)
	// 			cv.TV = epv.TV // TODO optimize? (epv.* ignored)
	// 		case *StructValue:
	// 			fpv := cb.GetPointerToInt(store, cv.Index)
	// 			cv.TV = fpv.TV // TODO optimize?
	// 		case *Block:
	// 			vpv := cb.GetPointerToInt(store, cv.Index)
	// 			cv.TV = vpv.TV // TODO optimize?

	// 		case *BoundMethodValue:
	// 			panic("should not happen: not a bound method")
	// 		case *MapValue:
	// 			panic("should not happen: not a map value")
	// 		default:
	// 			panic("should not happen")
	// 		}
	// 		tv.V = cv
	// 	}
	// do nothing
	// }

	// _, _ = oid, typeURL

	return nil
}

func (d decoder) unmarshalValue(tv *TypedValue) error {
	d.opts.DepthLimit--
	if d.opts.DepthLimit < 0 {
		return errors.New("exceeded max recursion depth")
	}

	// XXX: no type / undefined
	if tv.T == nil {
		// Read field name.
		tok, err := d.Read()
		if err != nil {
			return err
		}

		if tok.Kind() != json.Null {
			return d.unexpectedTokenError(tok)
		}

		return nil
	}

	switch tv.T.Kind() {
	case BoolKind, StringKind,
		IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind,
		UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind,
		Float32Kind, Float64Kind,
		BigintKind, BigdecKind:
		return d.unmarshalSingular(tv)

	case ArrayKind, SliceKind, TupleKind: // List
		return d.unmarshalListValue(tv)

	case StructKind:
		return d.unmarshalStructValue(tv)

		// case InterfaceKind:
		// 	return decoder.unmarshalAny

	case PointerKind:
		return d.unmarshalPointerValue(tv)

	}

	panic("not implemented")
}

// marshalScalar marshals the given non-repeated field value. This includes
// all scalar types, enums, messages, and groups.
func (e encoder) marshalScalar(tv *TypedValue) error {
	if len(tv.N) == 0 {
		e.WriteNull()
		return nil
	}

	switch kind := tv.T.Kind(); kind {
	case BoolKind:
		e.WriteBool(tv.GetBool())
	case StringKind:
		e.WriteString(tv.GetString())
	case IntKind:
		e.WriteInt(tv.GetInt())
	case Int8Kind:
		e.WriteInt8(tv.GetInt8())
	case Int16Kind:
		e.WriteInt16(tv.GetInt16())
	case Int32Kind:
		e.WriteInt32(tv.GetInt32())
	case Int64Kind:
		e.WriteInt64(tv.GetInt64())
	case UintKind:
		e.WriteUint(tv.GetUint())
	case Uint8Kind:
		e.WriteUint8(tv.GetUint8())
	case Uint16Kind:
		e.WriteUint16(tv.GetUint16())
	case Uint32Kind:
		e.WriteUint32(tv.GetUint32())
	case Uint64Kind:
		e.WriteUint64(tv.GetUint64())
	case Float32Kind:
		e.WriteFloat32(tv.GetFloat32())
	case Float64Kind:
		e.WriteFloat64(tv.GetFloat64())
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind.String()))
	}

	return nil
}

// marshalSingular marshals the given non-repeated field value. This includes
// all scalar types, enums, messages, and groups.
func (d decoder) unmarshalSingular(tv *TypedValue) error {
	tok, err := d.Read()
	if err != nil {
		return err
	}

	// XXX: guess unknown type
	// if tv.T == nil {
	// 	if !d.unmarshalUnknownNumber(tv, tok) {
	// 		return d.newError(tok.Pos(), "invalid value for %v field %v: %v", tok.Kind(), tok.Name(), tok.RawString())
	// 	}

	// 	return nil
	// }

	var ok bool
	switch kind := tv.T.Kind(); kind {
	case BoolKind:
		if ok = tok.Kind() == json.Bool; ok {
			tv.SetBool(tok.Bool())
		}
	case StringKind:
		if ok = tok.Kind() == json.String; ok {
			tv.SetString(StringValue(tok.ParsedString()))
		}
	case IntKind, Int16Kind, Int8Kind, Int32Kind, Int64Kind:
		ok = unmarshalInt(tv, tok)
	case UintKind, Uint16Kind, Uint8Kind, Uint32Kind, Uint64Kind:
		ok = unmarshalUint(tv, tok)
	case Float32Kind, Float64Kind:
		ok = unmarshalFloat(tv, tok)
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind.String()))
	}

	if ok {
		return nil
	}

	return d.newError(tok.Pos(), "invalid type %v for %v: %v",
		tv.T.String(), tok.Kind(), tok.RawString())
}

func getBitsize(t Type) int {
	switch k := t.Kind(); k {
	case Int8Kind, Uint8Kind:
		return 8
	case Int16Kind, Uint16Kind:
		return 16
	case Int32Kind, Float32Kind, Uint32Kind:
		return 32
	case UintKind, IntKind, Int64Kind, Uint64Kind, Float64Kind:
		return 64
	default:
		panic("cannot not guess no bitSize of type: " + k.String())
	}

}

// XXX: wip guess_number
// func (d *decoder) unmarshalUnknownNumber(tv *TypedValue, tok json.Token) bool {
// 	alloc := d.opts.Store.GetAllocator()
// 	switch tok.Kind() {
// 	case json.Number:
// 		if v, ok := tok.Int(64); ok {
// 			tv.T = alloc.NewType(IntType)
// 			tv.SetInt(int(v))
// 		} else if v, ok := tok.Uint(64); ok {
// 			tv.T = alloc.NewType(UintType)
// 			tv.SetUint(uint(v))
// 		} else if v, ok := tok.Float(64); ok {
// 			tv.T = alloc.NewType(Float64Type)
// 			tv.SetFloat64(v)
// 		} else {
// 			return false
// 		}

// 		return true

// 	case json.String:
// 		// Decode number from string.
// 		s := strings.TrimSpace(tok.ParsedString())
// 		if len(s) != len(tok.ParsedString()) {
// 			return false
// 		}
// 		dec := json.NewDecoder([]byte(s))
// 		tok, err := dec.Read()
// 		if err != nil {
// 			return false
// 		}

// 		return d.unmarshalUnknownNumber(tv, tok)
// 	}

// 	return false
// }

func unmarshalInt(tv *TypedValue, tok json.Token) bool {
	bitSize := getBitsize(tv.T)
	switch tok.Kind() {
	case json.Number:
		return setInt(tv, tok, bitSize)

	case json.String:
		// Decode number from string.
		s := strings.TrimSpace(tok.ParsedString())
		if len(s) != len(tok.ParsedString()) {
			return false
		}
		dec := json.NewDecoder([]byte(s))
		tok, err := dec.Read()
		if err != nil {

			return false
		}
		return setInt(tv, tok, bitSize)
	}

	return false
}

func setInt(tv *TypedValue, tok json.Token, bitSize int) bool {
	var ok bool
	var n int64

	switch bt := tv.T.Kind(); bt {
	case IntKind:
		if n, ok = tok.Int(bitSize); ok {
			tv.SetInt(int(n))
		}
	case Int8Kind:
		if n, ok = tok.Int(bitSize); ok {
			tv.SetInt8(int8(n))
		}
	case Int16Kind:
		if n, ok = tok.Int(bitSize); ok {
			tv.SetInt16(int16(n))
		}
	case Int32Kind:
		if n, ok = tok.Int(bitSize); ok {
			tv.SetInt32(int32(n))
		}
	case Int64Kind:
		if n, ok = tok.Int(bitSize); ok {
			tv.SetInt64(n)
		}
	default:
		panic(fmt.Sprintf("invalid int kind: %s", bt.String()))
	}

	return ok
}

func unmarshalUint(tv *TypedValue, tok json.Token) bool {
	bitSize := getBitsize(tv.T)
	switch tok.Kind() {
	case json.Number:
		return setUint(tv, tok, bitSize)

	case json.String:
		// Decode number from string.
		s := strings.TrimSpace(tok.ParsedString())
		if len(s) != len(tok.ParsedString()) {
			return false
		}
		dec := json.NewDecoder([]byte(s))
		tok, err := dec.Read()
		if err != nil {

			return false
		}
		return setUint(tv, tok, bitSize)
	}

	return false
}

func setUint(tv *TypedValue, tok json.Token, bitSize int) bool {
	var ok bool
	var n uint64

	switch bt := tv.T.Kind(); bt {
	case UintKind:
		if n, ok = tok.Uint(bitSize); ok {
			tv.SetUint(uint(n))
		}
	case Uint8Kind:
		if n, ok = tok.Uint(bitSize); ok {
			tv.SetUint8(uint8(n))
		}
	case Uint16Kind:
		if n, ok = tok.Uint(bitSize); ok {
			tv.SetUint16(uint16(n))
		}
	case Uint32Kind:
		if n, ok = tok.Uint(bitSize); ok {
			tv.SetUint32(uint32(n))
		}
	case Uint64Kind:
		if n, ok = tok.Uint(bitSize); ok {
			tv.SetUint64(n)
		}
	default:
		panic(fmt.Sprintf("invalid uint kind: %s", bt.String()))
	}

	return ok
}

func unmarshalFloat(tv *TypedValue, tok json.Token) bool {
	bitSize := getBitsize(tv.T)
	switch tok.Kind() {
	case json.Number:
		return setFloat(tv, tok, bitSize)

	case json.String:
		s := tok.ParsedString()

		// XXX: do we need to suport this ?
		// switch s {
		// case "NaN":
		// 	if bitSize == 32 {
		// 		tv.Set
		// 		return
		// 	}
		// 	return protoreflect.ValueOfFloat64(math.NaN()), true
		// case "Infinity":
		// 	if bitSize == 32 {
		// 		return protoreflect.ValueOfFloat32(float32(math.Inf(+1))), true
		// 	}
		// 	return protoreflect.ValueOfFloat64(math.Inf(+1)), true
		// case "-Infinity":
		// 	if bitSize == 32 {
		// 		return protoreflect.ValueOfFloat32(float32(math.Inf(-1))), true
		// 	}
		// 	return protoreflect.ValueOfFloat64(math.Inf(-1)), true
		// }

		// Decode number from string.
		if len(s) != len(strings.TrimSpace(s)) {
			return false
		}
		dec := json.NewDecoder([]byte(s))
		tok, err := dec.Read()
		if err != nil {
			return false
		}
		return setFloat(tv, tok, bitSize)
	}
	return false
}

func setFloat(tv *TypedValue, tok json.Token, bitSize int) bool {
	var ok bool
	var n float64

	switch bt := tv.T.Kind(); bt {
	case Float32Kind:
		if n, ok = tok.Float(bitSize); ok {
			tv.SetFloat32(float32(n))
		}
	case Float64Kind:
		if n, ok = tok.Float(bitSize); ok {
			tv.SetFloat64(float64(n))
		}
	default:
		panic(fmt.Sprintf("invalid uint kind: %s", bt.String()))
	}

	return ok
}

// XXX: The JSON representation of an Interface message uses the regular representation of
// the deserialized, embedded message, with an additional field `@type` which
// contains the type URL. If the embedded message type is well-known and has a
// custom JSON representation, that representation will be embedded adding a
// field `value` which holds the custom JSON in addition to the `@type` field.
func (e encoder) marshalInterfaceValue(tv *TypedValue) error {
	panic("no implemented")
}

// XXX: The JSON representation of an Interface message uses the regular representation of
// the deserialized, embedded message, with an additional field `@type` which
// contains the type URL. If the embedded message type is well-known and has a
// custom JSON representation, that representation will be embedded adding a
// field `value` which holds the custom JSON in addition to the `@type` field.
func (e decoder) unmarshalInterfaceValue(tv *TypedValue) error {
	panic("no implemented")
}

var ErrRecursivePointer = errors.New(`recursive detected`)
var ErrMissingType = errors.New(`missing "@type" field`)
var ErrEmptyObject = errors.New(`empty object`)

// Wrapper types are encoded as JSON primitives like string, number or boolean.
func (e encoder) marshalPointerValue(tv *TypedValue) error {
	if tv.V == nil {
		return nil
	}

	pv := tv.V.(PointerValue)
	o, ok := pv.Base.(Object)
	if !ok {
		panic(ErrEmptyObject)
	}

	id := o.GetObjectID()
	if e.cache[id.String()] {
		panic(ErrRecursivePointer)
	}
	e.cache[id.String()] = true

	etv := pv.Deref()
	return e.marshalValue(&etv)
}

func (d decoder) unmarshalPointerValue(tv *TypedValue) error {
	var eltv TypedValue
	eltv.T = tv.T.(*PointerType).Elem()
	if err := d.unmarshalValue(&eltv); err != nil {
		return err
	}

	// XXX: Alloc ?
	tv.V = PointerValue{TV: &eltv}
	return nil
}

func (e encoder) marshalStructValue(tv *TypedValue) error {
	e.StartObject()
	defer e.EndObject()

	// XXX: assert type/value ?
	st := baseOf(tv.T).(*StructType)
	sv := tv.V.(*StructValue)
	for i := range st.Fields {
		ft := st.Fields[i]

		name, opts, err := getFieldName(ft)
		if err != nil {
			return err
		}

		// Empty name mean either:
		// - unexported field
		// - tag is invalid
		// - tag is equal to "-"
		if name == "" {
			continue
		}

		fv := &sv.Fields[i]
		if opts.Contains("omitempty") && isEmptyValue(fv) {
			continue
		}

		fmt.Println(name)
		if name != "" {
			e.WriteName(name)
		} else {
			e.WriteName(string(ft.Name))
		}

		if err := e.marshalValue(fv); err != nil {
			return err
		}
	}

	return nil
}

func (d *decoder) unmarshalStructValue(tv *TypedValue) error {
	st := baseOf(tv.T).(*StructType)
	tok, err := d.Read()
	if err != nil {
		return err
	}
	if tok.Kind() != json.ObjectOpen {
		return d.unexpectedTokenError(tok)
	}

	// Allocate and construct index field reference
	fvs := d.opts.Alloc.NewStructFields(len(st.Fields))
	mfvs := make(map[string]int)
	for i, ft := range st.Fields {
		name, _, err := getFieldName(ft)
		if err != nil {
			return err
		}
		mfvs[name] = i
	}

	for {
		// Read field name.
		tok, err := d.Read()
		if err != nil {
			return err
		}

		switch tok.Kind() {
		default:
			return d.unexpectedTokenError(tok)
		case json.ObjectClose:
			tv.V = d.opts.Alloc.NewStruct(fvs)
			return nil
		case json.Name:
			// Continue below.
		}

		name := tok.Name()

		i, ok := mfvs[name]
		if !ok {
			return d.newError(tok.Pos(), "unknown field name %q", name)
		}

		fv := &fvs[i]
		fv.T = st.Fields[i].Type
		if err := d.unmarshalValue(fv); err != nil {
			return err
		}
	}
}

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

func getFieldName(ft FieldType) (string, tagOptions, error) {
	jsontag := ft.Tag.Get("json")
	tagname, opts, _ := parseTagValue(jsontag)
	if tagname == "-" {
		return "", "", nil
	}

	if !isValidTag(tagname) { // XXX: should this return an error instead ?
		tagname = ""
	}

	if !ft.IsExported() {
		if jsontag != "" {
			return "", "", fmt.Errorf("struct field %q has json tag but is not exported", ft.Name)
		}

		return "", "", nil
	}

	if tagname != "" {
		return tagname, opts, nil
	}

	return string(ft.Name), opts, nil
}

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTagValue(tag string) (string, tagOptions, bool) {
	tag, opt, ok := strings.Cut(tag, ",")
	return tag, tagOptions(opt), ok
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var name string
		name, s, _ = strings.Cut(s, ",")
		if name == optionName {
			return true
		}
	}
	return false
}

func isValidTag(s string) bool {
	// Reserve '@' prefix for special tag
	if s == "" || s[0] == '@' {
		return false
	}

	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			return false
		}
	}
	return true
}

func isEmptyValue(tv *TypedValue) bool {
	if tv.T == nil {
		return true
	}

	switch tv.T.Kind() {
	case ArrayKind, MapKind, SliceKind, StringKind:
		return tv.GetLength() == 0
	default:
		return tv.V == nil && tv.N == [8]byte{}
	}
}

// The JSON representation for ListValue is JSON array that contains the encoded
// ListValue.values repeated field and follows the serialization rules for a
// repeated field.
func (e encoder) marshalListValue(tv *TypedValue) error {
	e.StartArray()
	defer e.EndArray()

	if tv.V == nil {
		return nil
	}

	switch tv.T.Kind() {
	case ArrayKind:
		e.marshalArrayValue(tv)
	case SliceKind:
		e.marshalSliceValue(tv)
	default:
		return fmt.Errorf("unknown list type: %s", tv.T.String())
	}

	return nil
}

func (d decoder) unmarshalListValue(tv *TypedValue) error {
	fmt.Println("unmarshal start")

	tok, err := d.Read()
	if err != nil {
		return err
	}
	if tok.Kind() != json.ArrayOpen {
		return d.unexpectedTokenError(tok)
	}

	fmt.Println(tv.String())
	switch tv.T.Kind() {
	case ArrayKind:
		d.unmarshalArrayValue(tv)
	case SliceKind:
		d.unmarshalSliceValue(tv)
	// case TupleKind:
	// d.unmarshalTupleValue(tv)
	default:
		return fmt.Errorf("unknown list type: %s", tv.T.String())
	}

	return nil
}

func (e encoder) marshalArrayValue(tv *TypedValue) {
	av := tv.V.(*ArrayValue)
	if av.Data != nil {
		e.WriteBytesArrayValue(av.Data)
		return
	}

	// XXX: handle Uint8 as base64 string ?

	// General case.
	avl := len(av.List)
	for i := 0; i < avl; i++ {
		etv := &av.List[i]
		e.marshalValue(etv)
	}
}

func (d decoder) unmarshalArrayValue(tv *TypedValue) error {
	at := tv.T.(*ArrayType)
	list := make([]TypedValue, 0, at.Len)

	// XXX: handle base64 []byte

	// XXX: should we guess the size of the array and alloc before
	// unmarshling internal value ?
	for {
		tok, err := d.Peek()
		if err != nil {
			return err
		}

		if tok.Kind() == json.ArrayClose {
			d.Read() // discard close token

			// make the actual alloc
			d.opts.Alloc.AllocateListArray(int64(len(list)))
			tv.V = &ArrayValue{
				List: list,
			}
			return nil
		}

		var eltv TypedValue
		eltv.T = at.Elt // copy type element
		if err := d.unmarshalValue(&eltv); err != nil {
			return err
		}

		list = append(list, eltv)
	}
}

func (e encoder) marshalSliceValue(tv *TypedValue) {
	sv := tv.V.(*SliceValue)
	svo := sv.Offset
	svl := sv.Length
	var av *ArrayValue

	switch cv := sv.Base.(type) {
	case nil:
		return
	case RefValue:
		store := e.store()
		av = store.GetObject(cv.ObjectID).(*ArrayValue)
		sv.Base = av

	case *ArrayValue:
		av = cv
	default:
		panic("should not happen")
	}

	if av.Data != nil {
		e.WriteBytesArrayValue(av.Data[svo:])
		return
	}

	for i := svo; i < svo+svl; i++ {
		e.marshalValue(&av.List[i])
	}
}

func (d decoder) unmarshalSliceValue(tv *TypedValue) error {
	st := tv.T.(*SliceType)
	list := make([]TypedValue, 0)

	// XXX: handle data base64 []byte

	// XXX: should we guess the size of the array and alloc before
	// unmarshling internal value ?
	for {
		tok, err := d.Peek()
		if err != nil {
			return err
		}

		if tok.Kind() == json.ArrayClose {
			d.Read() // discard close token

			// make the actual alloc
			tv.V = d.opts.Alloc.NewSliceFromList(list)
			return nil
		}

		var eltv TypedValue
		eltv.T = st.Elt // copy type element

		fmt.Println("unarmshal val:", eltv.T.String())
		if err := d.unmarshalValue(&eltv); err != nil {
			return err
		}

		list = append(list, eltv)
	}
}

func (e encoder) WriteBytesArrayValue(data []byte) {
	e.WriteString(base64.StdEncoding.EncodeToString(data))
}

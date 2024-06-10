package gnolang

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/encoding/json"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const defaultIndent = "  "
const defaultRecursionLimit = 10000

// Format formats the message as a multiline string.
// This function is only intended for human consumption and ignores errors.
// Do not depend on the output being stable. Its output will change across
// different builds of your program, even when using the same version of the
// protobuf module.
// func Format(m *TypedValue) string {
// 	return MarshalOptions{Multiline: true}.Format(m)
// }

// Marshal writes the given [*TypedValue] in JSON format using default options.
// Do not depend on the output being stable. Its output will change across
// different builds of your program, even when using the same version of the
// protobuf module.
func (tv *TypedValue) Marshal() ([]byte, error) {
	return MarshalOptions{}.Marshal(tv)
}

// MarshalOptions is a configurable JSON format marshaler.
type MarshalOptions struct {
	// pragma.NoUnkeyedLiterals

	// Multiline specifies whether the marshaler should format the output in
	// indented-form with every textual element on a new line.
	// If Indent is an empty string, then an arbitrary indent is chosen.
	Multiline bool

	// Indent specifies the set of indentation characters to use in a multiline
	// formatted output such that every entry is preceded by Indent and
	// terminated by a newline. If non-empty, then Multiline is treated as true.
	// Indent can only be composed of space or tab characters.
	Indent string

	// AllowPartial allows messages that have missing required fields to marshal
	// without returning an error. If AllowPartial is false (the default),
	// Marshal will return error if there are any missing required fields.
	AllowPartial bool

	// UseProtoNames uses proto field name instead of lowerCamelCase name in JSON
	// field names.
	UseProtoNames bool

	// UseEnumNumbers emits enum values as numbers.
	UseEnumNumbers bool

	// EmitUnpopulated specifies whether to emit unpopulated fields. It does not
	// emit unpopulated oneof fields or unpopulated extension fields.
	// The JSON value emitted for unpopulated fields are as follows:
	//  ╔═══════╤════════════════════════════╗
	//  ║ JSON  │ Protobuf field             ║
	//  ╠═══════╪════════════════════════════╣
	//  ║ false │ proto3 boolean fields      ║
	//  ║ 0     │ proto3 numeric fields      ║
	//  ║ ""    │ proto3 string/bytes fields ║
	//  ║ null  │ proto2 scalar fields       ║
	//  ║ null  │ message fields             ║
	//  ║ []    │ list fields                ║
	//  ║ {}    │ map fields                 ║
	//  ╚═══════╧════════════════════════════╝
	EmitUnpopulated bool

	// EmitDefaultValues specifies whether to emit default-valued primitive fields,
	// empty lists, and empty maps. The fields affected are as follows:
	//  ╔═══════╤════════════════════════════════════════╗
	//  ║ JSON  │ Protobuf field                         ║
	//  ╠═══════╪════════════════════════════════════════╣
	//  ║ false │ non-optional scalar boolean fields     ║
	//  ║ 0     │ non-optional scalar numeric fields     ║
	//  ║ ""    │ non-optional scalar string/byte fields ║
	//  ║ []    │ empty repeated fields                  ║
	//  ║ {}    │ empty map fields                       ║
	//  ╚═══════╧════════════════════════════════════════╝
	//
	// Behaves similarly to EmitUnpopulated, but does not emit "null"-value fields,
	// i.e. presence-sensing fields that are omitted will remain omitted to preserve
	// presence-sensing.
	// EmitUnpopulated takes precedence over EmitDefaultValues since the former generates
	// a strict superset of the latter.
	EmitDefaultValues bool

	// Resolver is used for looking up types when expanding google.protobuf.Any
	// messages. If nil, this defaults to using protoregistry.GlobalTypes.
	// Resolver interface {
	// 	protoregistry.ExtensionTypeResolver
	// 	protoregistry.MessageTypeResolver
	// }

	Store Store
}

// UnmarshalOptions is a configurable JSON format parser.
type UnmarshalOptions struct {
	// If AllowPartial is set, input for messages that will result in missing
	// required fields will not return an error.
	AllowPartial bool

	// If DiscardUnknown is set, unknown fields and enum name values are ignored.
	DiscardUnknown bool

	// RecursionLimit limits how deeply messages may be nested.
	// If zero, a default limit is applied.
	RecursionLimit int

	Store Store
}

// Format formats the message as a string.
// This method is only intended for human consumption and ignores errors.
// Do not depend on the output being stable. Its output will change across
// different builds of your program, even when using the same version of the
// protobuf module.
// XXX: ignore me ?
// func (o MarshalOptions) Format(m *TypedValue) string {
// 	if m == nil || !m.ProtoReflect().IsValid() {
// 		return "<nil>" // invalid syntax, but okay since this is for debugging
// 	}
// 	o.AllowPartial = true
// 	b, _ := o.Marshal(m)
// 	return string(b)
// }

// Marshal marshals the given [*TypedValue] in the JSON format using options in
// Do not depend on the output being stable. Its output will change across
// different builds of your program, even when using the same version of the
// protobuf module.
func (o MarshalOptions) Marshal(tv *TypedValue) ([]byte, error) {
	return o.marshal(nil, tv)
}

// Unmarshal reads the given []byte and populates the given [proto.Message]
// using options in the UnmarshalOptions object.
// It will clear the message first before setting the fields.
// If it returns an error, the given message may be partially set.
// The provided message must be mutable (e.g., a non-nil pointer to a message).
func (o UnmarshalOptions) Unmarshal(b []byte, tv *TypedValue) error {
	return o.unmarshal(b, tv)
}

// unmarshal is a centralized function that all unmarshal operations go through.
// For profiling purposes, avoid changing the name of this function or
// introducing other code paths for unmarshal that do not go through this.
func (o UnmarshalOptions) unmarshal(b []byte, tv *TypedValue) error {
	// tv.Reset()  XXX: reset typed value ?

	// XXX: Set a default store
	// if o.Resolver == nil {
	// 	o.Resolver = protoregistry.GlobalTypes
	// }

	if o.RecursionLimit == 0 {
		o.RecursionLimit = defaultRecursionLimit
	}

	dec := decoder{json.NewDecoder(b), o}
	if err := dec.unmarshalMessage(tv, false); err != nil {
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
// For profiling purposes, avoid changing the name of this function or
// introducing other code paths for marshal that do not go through this.
func (o MarshalOptions) marshal(b []byte, tv *TypedValue) ([]byte, error) {
	if o.Multiline && o.Indent == "" {
		o.Indent = defaultIndent
	}

	// XXX: Use store as resolver
	if o.Store == nil {
		panic("no store has been set")
		// o.Resolver = protoregistry.GlobalTypes
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

	enc := encoder{internalEnc, o}
	if err := enc.marshalMessage(tv); err != nil {
		return nil, err
	}

	return enc.Bytes(), nil
}

type encoder struct {
	*json.Encoder
	opts MarshalOptions
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

type marshalFunc func(encoder, *TypedValue) error

// wellKnownTypeMarshaler returns a marshal function if the message type
// has specialized serialization behavior. It returns nil otherwise.
func wellKnownTypeMarshaler(tv *TypedValue) marshalFunc {
	switch tv.T.Kind() {
	case BoolKind, StringKind,
		IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind,
		UintKind, Uint8Kind, Uint16Kind, Uint64Kind,
		Float32Kind, Float64Kind,
		BigintKind, BigdecKind:
		return encoder.marshalSingular

	case StructKind:
		return encoder.marshalStructValue

	case ArrayKind, SliceKind, TupleKind: // List
		return encoder.marshalListValue

	case InterfaceKind:
		return encoder.marshalAny

	case PointerKind:
		return encoder.marshalPointerValue

	case RefTypeKind:
		return nil
	}

	return nil
}

type unmarshalFunc func(decoder, *TypedValue) error

// wellKnownTypeUnmarshaler returns a unmarshal function if the message type
// has specialized serialization behavior. It returns nil otherwise.
func wellKnownTypeUnmarshaler(tv *TypedValue) unmarshalFunc {
	switch tv.T.Kind() {
	case BoolKind, StringKind,
		IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind,
		UintKind, Uint8Kind, Uint16Kind, Uint64Kind,
		Float32Kind, Float64Kind,
		BigintKind, BigdecKind:
		return decoder.marshalSingular

	case StructKind:
		return decoder.marshalStructValue

	case ArrayKind, SliceKind, TupleKind: // List
		return decoder.marshalListValue

	case InterfaceKind:
		return decoder.marshalAny

	case PointerKind:
		return decoder.marshalPointerValue

	case RefTypeKind:
		return nil
	}

	return nil
}

// marshalMessage marshals the fields in the given protoreflect.Message.
// If the typeURL is non-empty, then a synthetic "@type" field is injected
// containing the URL as the value.
func (e encoder) marshalMessage(tv *TypedValue) error {
	if marshal := wellKnownTypeMarshaler(tv); marshal != nil {
		return marshal(e, tv)
	}

	e.StartObject()
	defer e.EndObject()

	if tv.V == nil {
		return nil
	}

	panic("no supported")
	// store := e.opts.Store

	// // var typeURL string
	// var oid string
	// switch cv := tv.V.(type) {
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
	// 		case *BoundMethodValue:
	// 			panic("should not happen: not a bound method")
	// 		case *MapValue:
	// 			panic("should not happen: not a map value")
	// 		case *Block:
	// 			vpv := cb.GetPointerToInt(store, cv.Index)
	// 			cv.TV = vpv.TV // TODO optimize?
	// 		default:
	// 			panic("should not happen")
	// 		}
	// 		tv.V = cv
	// 	}
	// default:
	// 	// do nothing
	// }

	return nil
}

// unmarshalMessage unmarshals a message into the given protoreflect.Message.
func (d decoder) unmarshalMessage(tv *TypedValue, skipTypeURL bool) error {
	d.opts.RecursionLimit--
	if d.opts.RecursionLimit < 0 {
		return errors.New("exceeded max recursion depth")
	}
	if unmarshal := wellKnownTypeUnmarshaler(tv); unmarshal != nil {
		return unmarshal(d, tv)
	}

	tok, err := d.Read()
	if err != nil {
		return err
	}
	if tok.Kind() != json.ObjectOpen {
		return d.unexpectedTokenError(tok)
	}

	// messageDesc := m.Descriptor()
	// if !flags.ProtoLegacy && messageset.IsMessageSet(messageDesc) {
	// 	return errors.New("no support for proto1 MessageSets")
	// }

	// var seenNums set.Ints
	// var seenOneofs set.Ints
	// fieldDescs := messageDesc.Fields()
	// for {
	// 	// Read field name.
	// 	tok, err := d.Read()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	switch tok.Kind() {
	// 	default:
	// 		return d.unexpectedTokenError(tok)
	// 	case json.ObjectClose:
	// 		return nil
	// 	case json.Name:
	// 		// Continue below.
	// 	}

	// 	name := tok.Name()
	// 	// Unmarshaling a non-custom embedded message in Any will contain the
	// 	// JSON field "@type" which should be skipped because it is not a field
	// 	// of the embedded message, but simply an artifact of the Any format.
	// 	if skipTypeURL && name == "@type" {
	// 		d.Read()
	// 		continue
	// 	}

	// 	// Get the FieldDescriptor.
	// 	var fd protoreflect.FieldDescriptor
	// 	if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
	// 		// Only extension names are in [name] format.
	// 		extName := protoreflect.FullName(name[1 : len(name)-1])
	// 		extType, err := d.opts.Resolver.FindExtensionByName(extName)
	// 		if err != nil && err != protoregistry.NotFound {
	// 			return d.newError(tok.Pos(), "unable to resolve %s: %v", tok.RawString(), err)
	// 		}
	// 		if extType != nil {
	// 			fd = extType.TypeDescriptor()
	// 			if !messageDesc.ExtensionRanges().Has(fd.Number()) || fd.ContainingMessage().FullName() != messageDesc.FullName() {
	// 				return d.newError(tok.Pos(), "message %v cannot be extended by %v", messageDesc.FullName(), fd.FullName())
	// 			}
	// 		}
	// 	} else {
	// 		// The name can either be the JSON name or the proto field name.
	// 		fd = fieldDescs.ByJSONName(name)
	// 		if fd == nil {
	// 			fd = fieldDescs.ByTextName(name)
	// 		}
	// 	}
	// 	if flags.ProtoLegacy {
	// 		if fd != nil && fd.IsWeak() && fd.Message().IsPlaceholder() {
	// 			fd = nil // reset since the weak reference is not linked in
	// 		}
	// 	}

	// 	if fd == nil {
	// 		// Field is unknown.
	// 		if d.opts.DiscardUnknown {
	// 			if err := d.skipJSONValue(); err != nil {
	// 				return err
	// 			}
	// 			continue
	// 		}
	// 		return d.newError(tok.Pos(), "unknown field %v", tok.RawString())
	// 	}

	// 	// Do not allow duplicate fields.
	// 	num := uint64(fd.Number())
	// 	if seenNums.Has(num) {
	// 		return d.newError(tok.Pos(), "duplicate field %v", tok.RawString())
	// 	}
	// 	seenNums.Set(num)

	// 	// No need to set values for JSON null unless the field type is
	// 	// google.protobuf.Value or google.protobuf.NullValue.
	// 	if tok, _ := d.Peek(); tok.Kind() == json.Null && !isKnownValue(fd) && !isNullValue(fd) {
	// 		d.Read()
	// 		continue
	// 	}

	// 	switch {
	// 	case fd.IsList():
	// 		list := m.Mutable(fd).List()
	// 		if err := d.unmarshalList(list, fd); err != nil {
	// 			return err
	// 		}
	// 	case fd.IsMap():
	// 		mmap := m.Mutable(fd).Map()
	// 		if err := d.unmarshalMap(mmap, fd); err != nil {
	// 			return err
	// 		}
	// 	default:
	// 		// If field is a oneof, check if it has already been set.
	// 		if od := fd.ContainingOneof(); od != nil {
	// 			idx := uint64(od.Index())
	// 			if seenOneofs.Has(idx) {
	// 				return d.newError(tok.Pos(), "error parsing %s, oneof %v is already set", tok.RawString(), od.FullName())
	// 			}
	// 			seenOneofs.Set(idx)
	// 		}

	// 		// Required or optional fields.
	// 		if err := d.unmarshalSingular(m, fd); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	return nil
}

// marshalSingular marshals the given non-repeated field value. This includes
// all scalar types, enums, messages, and groups.
func (e encoder) marshalSingular(tv *TypedValue) error {
	if tv.V == nil {
		e.WriteNull()
		return nil
	}

	switch kind := tv.T.Kind(); kind {
	case BoolKind:
		e.WriteBool(tv.GetBool())
	case StringKind:
		e.WriteString(tv.String())
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

	// case protoreflect.MessageKind, protoreflect.GroupKind:
	// 	if err := e.marshalMessage(val.Message(), ""); err != nil {
	// 		return err
	// 	}

	return nil
}

// marshalSingular marshals the given non-repeated field value. This includes
// all scalar types, enums, messages, and groups.
func (d decoder) unmarshalSingular(tv *TypedValue) error {
	tok, err := d.Read()
	if err != nil {
		return err
	}

	switch kind := tv.T.Kind(); kind {
	case BoolKind:
		if tok.Kind() == json.Bool {
			tv.SetBool(tok.Bool())
		}
	case StringKind:
		if tok.Kind() == json.String {
			tv.SetString(StringValue(tok.ParsedString()))
		}
	case IntKind:
		d.opts.Store.GetAllocator().NewNative()
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

	// case protoreflect.MessageKind, protoreflect.GroupKind:
	// 	if err := e.marshalMessage(val.Message(), ""); err != nil {
	// 		return err
	// 	}

	return nil
}

func unmarshalInt(tok json.Token, bitSize int) (Value, bool) {
	switch tok.Kind() {
	case json.Number:
		return getInt(tok, bitSize)

	case json.String:
		// Decode number from string.
		s := strings.TrimSpace(tok.ParsedString())
		if len(s) != len(tok.ParsedString()) {
			return protoreflect.Value{}, false
		}
		dec := json.NewDecoder([]byte(s))
		tok, err := dec.Read()
		if err != nil {
			return protoreflect.Value{}, false
		}
		return getInt(tok, bitSize)
	}

	return nil, false
}

func getInt(tok json.Token, bitSize int) (Value, bool) {
	n, ok := tok.Int(bitSize)
	if !ok {
		return protoreflect.Value{}, false
	}
	if bitSize == 32 {
		return protoreflect.ValueOfInt32(int32(n)), true
	}
	return protoreflect.ValueOfInt64(n), true
}

func unmarshalUint(tok json.Token, bitSize int) (protoreflect.Value, bool) {
	switch tok.Kind() {
	case json.Number:
		return getUint(tok, bitSize)

	case json.String:
		// Decode number from string.
		s := strings.TrimSpace(tok.ParsedString())
		if len(s) != len(tok.ParsedString()) {
			return protoreflect.Value{}, false
		}
		dec := json.NewDecoder([]byte(s))
		tok, err := dec.Read()
		if err != nil {
			return protoreflect.Value{}, false
		}
		return getUint(tok, bitSize)
	}
	return protoreflect.Value{}, false
}

func getUint(tok json.Token, bitSize int) (protoreflect.Value, bool) {
	n, ok := tok.Uint(bitSize)
	if !ok {
		return protoreflect.Value{}, false
	}

	if bitSize == 32 {
		return protoreflect.ValueOfUint32(uint32(n)), true
	}
	return protoreflect.ValueOfUint64(n), true
}

func unmarshalFloat(tok json.Token, bitSize int) (protoreflect.Value, bool) {
	switch tok.Kind() {
	case json.Number:
		return getFloat(tok, bitSize)

	case json.String:
		s := tok.ParsedString()
		switch s {
		case "NaN":
			if bitSize == 32 {
				return float32(math.NaN())), true
			}
			return protoreflect.ValueOfFloat64(math.NaN()), true
		case "Infinity":
			if bitSize == 32 {
				return protoreflect.ValueOfFloat32(float32(math.Inf(+1))), true
			}
			return protoreflect.ValueOfFloat64(math.Inf(+1)), true
		case "-Infinity":
			if bitSize == 32 {
				return protoreflect.ValueOfFloat32(float32(math.Inf(-1))), true
			}
			return protoreflect.ValueOfFloat64(math.Inf(-1)), true
		}

		// Decode number from string.
		if len(s) != len(strings.TrimSpace(s)) {
			return protoreflect.Value{}, false
		}
		dec := json.NewDecoder([]byte(s))
		tok, err := dec.Read()
		if err != nil {
			return protoreflect.Value{}, false
		}
		return getFloat(tok, bitSize)
	}
	return protoreflect.Value{}, false
}

func getFloat(tok json.Token, bitSize int) (protoreflect.Value, bool) {
	n, ok := tok.Float(bitSize)
	if !ok {
		return protoreflect.Value{}, false
	}
	if bitSize == 32 {
		return protoreflect.ValueOfFloat32(float32(n)), true
	}
	return protoreflect.ValueOfFloat64(n), true
}

// The JSON representation of an Any message uses the regular representation of
// the deserialized, embedded message, with an additional field `@type` which
// contains the type URL. If the embedded message type is well-known and has a
// custom JSON representation, that representation will be embedded adding a
// field `value` which holds the custom JSON in addition to the `@type` field.
func (e encoder) marshalAny(tv *TypedValue) error {
	return fmt.Errorf("any: TODO")
}

var errEmptyObject = fmt.Errorf(`empty object`)
var errMissingType = fmt.Errorf(`missing "@type" field`)

// Wrapper types are encoded as JSON primitives like string, number or boolean.

func (e encoder) marshalPointerValue(tv *TypedValue) error {
	pv := tv.V.(PointerValue)
	etv := pv.TV
	return e.marshalMessage(etv)
}

func (e encoder) marshalStructValue(tv *TypedValue) error {
	e.StartObject()
	defer e.EndObject()

	// XXX: assert type/value ?
	st := baseOf(tv.T).(*StructType)
	sv := tv.V.(*StructValue)
	for i := range st.Fields {
		ft := st.Fields[i]
		jsontag := ft.Tag.Get("json")

		if !ft.IsExported() {
			if jsontag != "" {
				return fmt.Errorf("struct field %s has json tag but is not exported", ft.Name)
			}

			continue
		}

		fv := &sv.Fields[i]
		if _, omitempty := ft.Tag.Lookup("omitempty"); omitempty && isEmptyValue(fv) {
			continue
		}

		if jsontag != "" {
			e.WriteName(string(ft.Tag))
		} else {
			e.WriteName(string(ft.Name))
		}

		e.marshalSingular(fv)
	}

	return nil
}

func isEmptyValue(tv *TypedValue) bool {
	switch tv.T.Kind() {
	case ArrayKind, MapKind, SliceKind, StringKind:
		return tv.GetLength() == 0
	default:
		return tv.V == nil
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
		e.marshalMessage(etv)
	}
}

func (e encoder) marshalSliceValue(tv *TypedValue) {
	sv := tv.V.(*SliceValue)
	svo := sv.Offset
	svl := sv.Length
	av := sv.GetBase(e.opts.Store)
	if av.Data != nil {
		e.WriteBytesArrayValue(av.Data[svo:])
		return
	}

	for i := svo; i < svo+svl; i++ {
		e.marshalSingular(&av.List[i])
	}
}

func (e encoder) WriteBytesArrayValue(data []byte) {
	e.WriteString(base64.StdEncoding.EncodeToString(data))
}

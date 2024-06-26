package gnolang

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/encoding/json"
)

const defaultIndent = "  "
const defaultRecursionLimit = 10000

// Format formats the message as a multiline string.
// This function is only intended for human consumption and ignores errors.
// Do not depend on the output being stable.
// func Format(m *TypedValue) string {
// 	return MarshalOptions{Multiline: true}.Format(m)
// }

// Marshal writes the given [*TypedValue] in JSON format using default options.
// Do not depend on the output being stable. Its output will change across
// different builds of your program, even when using the same version of the
// protobuf module.
func (tv *TypedValue) Marshal() ([]byte, error) {
	return MarshalOptions{
		Store: nil,
	}.Marshal(tv)
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

	// Resolver is used for looking up types when expanding google.protobuf.Any
	// messages. If nil, this defaults to using protoregistry.GlobalTypes.
	// Resolver interface {
	// 	protoregistry.ExtensionTypeResolver
	// 	protoregistry.MessageTypeResolver
	// }

	Store Store

	Alloc *Allocator
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

	Alloc *Allocator
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
	if o.Alloc == nil {
		o.Alloc = nilAllocator
	}

	if o.Store == nil {
		o.Store = NewStore(o.Alloc, nil, nil)
	}

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

	if o.Alloc == nil {
		o.Alloc = nilAllocator
	}

	if o.Store == nil {
		o.Store = NewStore(o.Alloc, nil, nil)
	}

	// XXX: Use store as resolver
	// if o.Store == nil {
	// 	panic("no store has been set")
	// 	// o.Resolver = protoregistry.GlobalTypes
	// }

	var buff bytes.Buffer
	internalEnc, err := json.NewEncoder(b, &buff, o.Indent)
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

	return buff.Bytes(), nil
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

type marshalFunc func(encoder, *TypedValue) error

// wellKnownTypeMarshaler returns a marshal function if the message type
// has specialized serialization behavior. It returns nil otherwise.
func wellKnownTypeMarshaler(tv *TypedValue) marshalFunc {
	switch tv.T.Kind() {
	case BoolKind, StringKind,
		IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind,
		UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind,
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
		UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind,
		Float32Kind, Float64Kind,
		BigintKind, BigdecKind:
		return decoder.unmarshalSingular

	// case StructKind:
	// 	return decoder.unmarshalStructValue

	// case ArrayKind, SliceKind, TupleKind: // List
	// 	return decoder.unmarshalListValue

	// case InterfaceKind:
	// 	return decoder.unmarshalAny

	// case PointerKind:
	// 	return decoder.unmarshalPointerValue

	case RefTypeKind:
		return nil
	}

	return nil
}

// marshalValue marshals the fields in the given TypedValue.
// If the typeURL is non-empty, then a synthetic "@type" field is injected
// containing the URL as the value.
func (e encoder) marshalValue(tv *TypedValue) error {
	if marshal := wellKnownTypeMarshaler(tv); marshal != nil {
		return marshal(e, tv)
	}

	e.StartObject()
	defer e.EndObject()

	if tv.V == nil {
		return nil
	}

	fmt.Printf("%+#v\n", tv.V)
	// store := e.opts.Store

	var typeURL string
	var oid string
	// switch cv := tv.V.(type) {
	// case TypeValue:

	// }

	v := copyValueWithRefs(tv.V)
	fmt.Printf("%#v\n", v)

	if ctv, ok := tv.V.(TypeValue); ok {
		switch cv := ctv.Type.(type) {
		case *DeclaredType:

			fmt.Println(cv)
		}
	}

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

	_, _ = oid, typeURL

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

	// XXX: guess unknown type
	// if tv.T == nil {
	// 	if !d.unmarshalUnknownNumber(tv, tok) {
	// 		return d.newError(tok.Pos(), "invalid value for %v field %v: %v", tok.Kind(), tok.Name(), tok.RawString())
	// 	}

	// 	return nil
	// }

	switch kind := tv.T.Kind(); kind {
	case BoolKind:
		if tok.Kind() == json.Bool {
			tv.SetBool(tok.Bool())
		}
	case StringKind:
		if tok.Kind() == json.String {
			tv.SetString(StringValue(tok.ParsedString()))
		}
	case IntKind, Int16Kind, Int8Kind, Int32Kind, Int64Kind:
		if ok := unmarshalInt(tv, tok); ok {
			return nil
		}
	case UintKind, Uint16Kind, Uint8Kind, Uint32Kind, Uint64Kind:
		if ok := unmarshalUint(tv, tok); ok {
			return nil
		}
	case Float32Kind, Float64Kind:
		if ok := unmarshalFloat(tv, tok); ok {
			return nil
		}
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind.String()))
	}

	return nil
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
	case Int32Kind:
		if n, ok = tok.Int(bitSize); ok {
			tv.SetInt32(int32(n))
		}
	case Int16Kind:
		if n, ok = tok.Int(bitSize); ok {
			tv.SetInt16(int16(n))
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
		// XXX: do we need to suuport this
		s := tok.ParsedString()
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

// The JSON representation of an Any message uses the regular representation of
// the deserialized, embedded message, with an additional field `@type` which
// contains the type URL. If the embedded message type is well-known and has a
// custom JSON representation, that representation will be embedded adding a
// field `value` which holds the custom JSON in addition to the `@type` field.
func (e encoder) marshalAny(tv *TypedValue) error {
	return fmt.Errorf("any: TODO")
}

var ErrRecursivePointer = errors.New(`recursive detected`)
var ErrMissingType = errors.New(`missing "@type" field`)
var ErrEmptyObject = errors.New(`empty object`)

// Wrapper types are encoded as JSON primitives like string, number or boolean.

func (e encoder) marshalPointerValue(tv *TypedValue) error {
	pv := tv.V.(PointerValue)
	o, ok := pv.Base.(Object)
	if !ok {
		panic(ErrEmptyObject)
	}

	if e.store() == nil {
		id := o.GetObjectID()
		if e.cache[id.String()] {
			panic(ErrRecursivePointer)
		}
		e.cache[id.String()] = true

		etv := pv.Deref()
		return e.marshalValue(&etv)
	}

	panic("not supported")
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
			e.WriteName(jsontag)
		} else {
			e.WriteName(string(ft.Name))
		}

		if err := e.marshalValue(fv); err != nil {
			return err
		}
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
		e.marshalValue(etv)
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
		if store == nil {
			return // cannot guess rev without a store
		}

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
		e.marshalSingular(&av.List[i])
	}
}

func (e encoder) WriteBytesArrayValue(data []byte) {
	e.WriteString(base64.StdEncoding.EncodeToString(data))
}

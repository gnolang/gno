package gnolang

import (
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/encoding/json"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const defaultIndent = "  "

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
	if err := enc.marshalMessage(tv, ""); err != nil {
		return nil, err
	}

	return enc.Bytes(), nil
}

type encoder struct {
	*json.Encoder
	opts MarshalOptions
}

// typeFieldDesc is a synthetic field descriptor used for the "@type" field.
// var typeFieldDesc = func() protoreflect.FieldDescriptor {
// 	var fd filedesc.Field
// 	fd.L0.FullName = "@type"
// 	fd.L0.Index = -1
// 	fd.L1.Cardinality = protoreflect.Optional
// 	fd.L1.Kind = protoreflect.StringKind
// 	return &fd
// }()

// typeURLFieldRanger wraps a protoreflect.Message and modifies its Range method
// to additionally iterate over a synthetic field for the type URL.
// type typeURLFieldRanger struct {
// 	order.FieldRanger
// 	typeURL string
// }

// func (m typeURLFieldRanger) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
// 	if !f(typeFieldDesc, protoreflect.ValueOfString(m.typeURL)) {
// 		return
// 	}
// 	m.FieldRanger.Range(f)
// }

// // unpopulatedFieldRanger wraps a protoreflect.Message and modifies its Range
// // method to additionally iterate over unpopulated fields.
// type unpopulatedFieldRanger struct {
// 	protoreflect.Message

// 	skipNull bool
// }

// func (m unpopulatedFieldRanger) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
// 	fds := m.Descriptor().Fields()
// 	for i := 0; i < fds.Len(); i++ {
// 		fd := fds.Get(i)
// 		if m.Has(fd) || fd.ContainingOneof() != nil {
// 			continue // ignore populated fields and fields within a oneofs
// 		}

// 		v := m.Get(fd)
// 		isProto2Scalar := fd.Syntax() == protoreflect.Proto2 && fd.Default().IsValid()
// 		isSingularMessage := fd.Cardinality() != protoreflect.Repeated && fd.Message() != nil
// 		if isProto2Scalar || isSingularMessage {
// 			if m.skipNull {
// 				continue
// 			}
// 			v = protoreflect.Value{} // use invalid value to emit null
// 		}
// 		if !f(fd, v) {
// 			return
// 		}
// 	}
// 	m.Message.Range(f)
// }

// marshalMessage marshals the fields in the given protoreflect.Message.
// If the typeURL is non-empty, then a synthetic "@type" field is injected
// containing the URL as the value.
func (e encoder) marshalMessage(tv *TypedValue, typeURL string) error {
	if marshal := wellKnownTypeMarshaler(tv); marshal != nil {
		return marshal(e, tv)
	}

	e.StartObject()
	defer e.EndObject()

	// var fields order.FieldRanger = m
	// switch {
	// case e.opts.EmitUnpopulated:
	// 	fields = unpopulatedFieldRanger{Message: m, skipNull: false}
	// case e.opts.EmitDefaultValues:
	// 	fields = unpopulatedFieldRanger{Message: m, skipNull: true}
	// }
	// if typeURL != "" {
	// 	fields = typeURLFieldRanger{fields, typeURL}
	// }

	// XXX: iterator through structure fields ?
	// var err error
	// order.RangeFields(fields, order.IndexNameFieldOrder, func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
	// 	name := fd.JSONName()
	// 	if e.opts.UseProtoNames {
	// 		name = fd.TextName()
	// 	}

	// 	if err = e.WriteName(name); err != nil {
	// 		return false
	// 	}
	// 	if err = e.marshalValue(v, fd); err != nil {
	// 		return false
	// 	}
	// 	return true
	// })
	return err
}

// marshalValue marshals the given protoreflect.Value.
func (e encoder) marshalValue(tv *TypedValue) error {
	switch {
	case tv.IsList():
		return e.marshalList(val.List(), fd)
	case fd.IsMap():
		return e.marshalMap(val.Map(), fd)
	default:
		return e.marshalSingular(val, fd)
	}
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

// marshalList marshals the given protoreflect.List.
func (e encoder) marshalList(list protoreflect.List, fd protoreflect.FieldDescriptor) error {
	e.StartArray()
	defer e.EndArray()

	for i := 0; i < list.Len(); i++ {
		item := list.Get(i)
		if err := e.marshalSingular(item, fd); err != nil {
			return err
		}
	}
	return nil
}

// marshalMap marshals given protoreflect.Map.
func (e encoder) marshalMap(mmap protoreflect.Map, fd protoreflect.FieldDescriptor) error {
	e.StartObject()
	defer e.EndObject()

	var err error
	order.RangeEntries(mmap, order.GenericKeyOrder, func(k protoreflect.MapKey, v protoreflect.Value) bool {
		if err = e.WriteName(k.String()); err != nil {
			return false
		}
		if err = e.marshalSingular(v, fd.MapValue()); err != nil {
			return false
		}
		return true
	})
	return err
}

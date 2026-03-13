package genproto2

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// === Entry Point ===

func (ctx *P3Context2) generateMarshal(sb *strings.Builder, info *amino.TypeInfo) error {
	tname := typeName(info)
	if tname == "" {
		return nil // skip anonymous types
	}
	// Skip non-struct non-AminoMarshaler types.
	if info.Type.Kind() != reflect.Struct && !info.IsAminoMarshaler {
		return nil
	}

	sb.WriteString(fmt.Sprintf("func (goo %s) MarshalBinary2(cdc *amino.Codec, w io.Writer) error {\n", tname))

	// Handle AminoMarshaler: convert to repr, then marshal repr.
	if info.IsAminoMarshaler {
		sb.WriteString("\trepr, err := goo.MarshalAmino()\n")
		sb.WriteString("\tif err != nil {\n\t\treturn err\n\t}\n")
		rinfo := info.ReprType
		if err := ctx.writeReprMarshal(sb, rinfo); err != nil {
			return err
		}
		sb.WriteString("\treturn nil\n")
		sb.WriteString("}\n\n")
		return nil
	}

	if err := ctx.writeStructMarshalBody(sb, info, "goo"); err != nil {
		return err
	}

	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
	return nil
}

// === AminoMarshaler Repr Handling ===

// writeReprMarshal handles marshaling via AminoMarshaler repr types.
// The generated MarshalBinary2 is called within a struct context (the original
// type IS a struct), so the repr is encoded bare — just like encodeReflectBinary
// would encode the repr after converting from the AminoMarshaler.
//
// For struct reprs: encode fields directly.
// For slice reprs: encode the packed or unpacked list directly (bare).
// For primitive reprs: the original type is a struct, so amino wraps
//   non-struct reprs via writeFieldIfNotEmpty with field 1 in the implicit struct.
func (ctx *P3Context2) writeReprMarshal(sb *strings.Builder, rinfo *amino.TypeInfo) error {
	rt := rinfo.Type
	fopts := amino.FieldOptions{}

	switch {
	case rt.Kind() == reflect.Struct:
		return ctx.writeStructMarshalBody(sb, rinfo, "repr")

	case isListType(rt):
		if !rinfo.IsStructOrUnpacked(fopts) {
			// Packed slice repr: wrap in implicit struct field 1, like a primitive.
			return ctx.writePackedSliceReprMarshal(sb, rinfo)
		}
		return ctx.writeSliceReprMarshal(sb, rinfo)

	default:
		// Primitive repr: the original type is a struct, so MarshalReflect
		// at the top level wraps it in an implicit struct (field 1 + value).
		// Since the caller of MarshalBinary2 expects bare struct encoding,
		// we do the same: write field 1 key + value.
		typ3 := rinfo.GetTyp3(fopts)
		sb.WriteString("\t{\n")
		sb.WriteString("\t\tvar buf bytes.Buffer\n")
		sb.WriteString(fmt.Sprintf("\t\tif err := amino.EncodeFieldNumberAndTyp3(&buf, 1, %s); err != nil {\n\t\t\treturn err\n\t\t}\n",
			typ3GoStr(typ3)))
		sb.WriteString("\t\tlBeforeValue := buf.Len()\n")

		switch rt.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sb.WriteString("\t\tif err := amino.EncodeVarint(&buf, int64(repr)); err != nil {\n\t\t\treturn err\n\t\t}\n")
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			sb.WriteString("\t\tif err := amino.EncodeUvarint(&buf, uint64(repr)); err != nil {\n\t\t\treturn err\n\t\t}\n")
		case reflect.String:
			sb.WriteString("\t\tif err := amino.EncodeString(&buf, string(repr)); err != nil {\n\t\t\treturn err\n\t\t}\n")
		case reflect.Bool:
			sb.WriteString("\t\tif err := amino.EncodeBool(&buf, bool(repr)); err != nil {\n\t\t\treturn err\n\t\t}\n")
		default:
			sb.WriteString("\t}\n")
			return fmt.Errorf("unsupported primitive repr type kind %v for AminoMarshaler", rt.Kind())
		}

		// Match writeFieldIfNotEmpty: skip if value is just 0x00.
		sb.WriteString("\t\tbz := buf.Bytes()\n")
		sb.WriteString("\t\tif buf.Len() == lBeforeValue+1 && bz[len(bz)-1] == 0x00 {\n")
		sb.WriteString("\t\t\t// skip empty\n")
		sb.WriteString("\t\t} else {\n")
		sb.WriteString("\t\t\tif _, err := w.Write(bz); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
	}

	return nil
}

// writePackedSliceReprMarshal handles packed slice reprs that need implicit struct
// field 1 wrapping (slice where IsStructOrUnpacked is false).
func (ctx *P3Context2) writePackedSliceReprMarshal(sb *strings.Builder, info *amino.TypeInfo) error {
	einfo := info.Elem
	if einfo == nil {
		return fmt.Errorf("slice type %v has nil Elem info", info.Type)
	}

	fopts := amino.FieldOptions{}
	beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8
	sb.WriteString("\tif len(repr) > 0 {\n")
	sb.WriteString("\t\tvar buf bytes.Buffer\n")
	sb.WriteString("\t\tfor _, elem := range repr {\n")
	if beOptionByte {
		// Element repr is uint8: encode as raw byte (not uvarint).
		if einfo.IsAminoMarshaler {
			sb.WriteString("\t\t\telemRepr, err := elem.MarshalAmino()\n")
			sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
			sb.WriteString("\t\t\tif err := amino.EncodeByte(&buf, uint8(elemRepr)); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
		} else {
			sb.WriteString("\t\t\tif err := amino.EncodeByte(&buf, uint8(elem)); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
		}
	} else if einfo.IsAminoMarshaler {
		sb.WriteString("\t\t\telemRepr, err := elem.MarshalAmino()\n")
		sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
		ctx.writeValueEncode(sb, "\t\t\t", "elemRepr", einfo.ReprType, fopts)
	} else {
		ctx.writeValueEncode(sb, "\t\t\t", "elem", einfo, fopts)
	}
	sb.WriteString("\t\t}\n")
	// Wrap in field 1 key + EncodeByteSlice.
	sb.WriteString("\t\tvar obuf bytes.Buffer\n")
	sb.WriteString("\t\tif err := amino.EncodeFieldNumberAndTyp3(&obuf, 1, amino.Typ3ByteLength); err != nil {\n\t\t\treturn err\n\t\t}\n")
	sb.WriteString("\t\tif err := amino.EncodeByteSlice(&obuf, buf.Bytes()); err != nil {\n\t\t\treturn err\n\t\t}\n")
	sb.WriteString("\t\tif _, err := w.Write(obuf.Bytes()); err != nil {\n\t\t\treturn err\n\t\t}\n")
	sb.WriteString("\t}\n")
	return nil
}

// writeSliceReprMarshal marshals a slice repr value (from AminoMarshaler).
// The slice is encoded bare (no outer length prefix), just like a struct's list field.
func (ctx *P3Context2) writeSliceReprMarshal(sb *strings.Builder, info *amino.TypeInfo) error {
	einfo := info.Elem
	if einfo == nil {
		return fmt.Errorf("slice type %v has nil Elem info", info.Type)
	}

	// Check if elements are ByteLength (unpacked) or packed.
	fopts := amino.FieldOptions{} // default
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength {
		// Packed form: single length-prefixed block.
		sb.WriteString("\tif len(repr) > 0 {\n")
		sb.WriteString("\t\tvar buf bytes.Buffer\n")
		sb.WriteString("\t\tfor _, elem := range repr {\n")
		if einfo.IsAminoMarshaler {
			// Convert element to repr before encoding.
			sb.WriteString("\t\t\telemRepr, err := elem.MarshalAmino()\n")
			sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
			ctx.writeValueEncode(sb, "\t\t\t", "elemRepr", einfo.ReprType, fopts)
		} else {
			ctx.writeValueEncode(sb, "\t\t\t", "elem", einfo, fopts)
		}
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tif err := amino.EncodeByteSlice(w, buf.Bytes()); err != nil {\n\t\t\treturn err\n\t\t}\n")
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field entries. For top-level repr slices
		// this is encoded with field number 1 per element.
		sb.WriteString("\tfor _, elem := range repr {\n")
		sb.WriteString(fmt.Sprintf("\t\tif err := amino.EncodeFieldNumberAndTyp3(w, 1, amino.Typ3ByteLength); err != nil {\n\t\t\treturn err\n\t\t}\n"))
		if einfo.Type.Kind() == reflect.Struct {
			// Struct element: encode bare to buf, then length-prefix.
			sb.WriteString("\t\tvar buf bytes.Buffer\n")
			if einfo.IsAminoMarshaler {
				sb.WriteString("\t\telemRepr, err := elem.MarshalAmino()\n")
				sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
				sb.WriteString("\t\t// encode elemRepr struct to buf\n")
				ctx.writeStructMarshalBodyInline(sb, einfo.ReprType, "elemRepr", "\t\t", "&buf")
			} else {
				sb.WriteString("\t\tif err := elem.MarshalBinary2(cdc, &buf); err != nil {\n\t\t\treturn err\n\t\t}\n")
			}
			sb.WriteString("\t\tif err := amino.EncodeByteSlice(w, buf.Bytes()); err != nil {\n\t\t\treturn err\n\t\t}\n")
		} else {
			// Non-struct ByteLength element: encode directly to w
			// (encode functions include own length prefix).
			ctx.writePrimitiveEncode(sb, "elem", einfo, fopts, "\t\t", "w")
		}
		sb.WriteString("\t}\n")
	}
	return nil
}

// === Struct Fields ===

// writeStructMarshalBody writes the field-by-field marshal code for a struct.
func (ctx *P3Context2) writeStructMarshalBody(sb *strings.Builder, info *amino.TypeInfo, recv string) error {
	for _, field := range info.Fields {
		if err := ctx.writeFieldMarshal(sb, field, recv); err != nil {
			return err
		}
	}
	return nil
}

// writeStructMarshalBodyInline writes struct marshal code targeting a specific writer variable.
func (ctx *P3Context2) writeStructMarshalBodyInline(sb *strings.Builder, info *amino.TypeInfo, recv, indent, writerVar string) {
	for _, field := range info.Fields {
		ctx.writeFieldMarshalInline(sb, field, recv, indent, writerVar)
	}
}

func (ctx *P3Context2) writeFieldMarshal(sb *strings.Builder, field amino.FieldInfo, recv string) error {
	finfo := field.TypeInfo
	fname := field.Name
	fnum := field.BinFieldNum
	fopts := field.FieldOptions
	ftype := field.Type
	isPtr := ftype.Kind() == reflect.Ptr

	accessor := fmt.Sprintf("%s.%s", recv, fname)

	// Handle unpacked lists.
	if field.UnpackedList {
		return ctx.writeUnpackedListMarshal(sb, accessor, finfo, fopts)
	}

	// Handle pointer fields.
	// Amino's isNonstructDefaultValue recurses through pointers: a non-nil
	// pointer to a zero/default value is still skipped (unless WriteEmpty).
	if isPtr {
		sb.WriteString(fmt.Sprintf("\tif %s != nil {\n", accessor))
		derefAccessor := fmt.Sprintf("(*%s)", accessor)
		if !field.WriteEmpty {
			zeroCheck := ctx.zeroCheck(derefAccessor, finfo, fopts)
			if zeroCheck != "" {
				sb.WriteString(fmt.Sprintf("\t\tif %s {\n", zeroCheck))
				ctx.writeFieldValueMarshal(sb, derefAccessor, fnum, finfo, fopts, true, "\t\t\t")
				sb.WriteString("\t\t}\n")
			} else {
				ctx.writeFieldValueMarshal(sb, derefAccessor, fnum, finfo, fopts, true, "\t\t")
			}
		} else {
			ctx.writeFieldValueMarshal(sb, derefAccessor, fnum, finfo, fopts, true, "\t\t")
		}
		sb.WriteString("\t}\n")
		return nil
	}

	// Handle AminoMarshaler fields: convert to repr, then encode repr.
	if finfo.IsAminoMarshaler && finfo.Type.Kind() != reflect.Struct {
		// Non-struct AminoMarshaler (e.g. crypto.Address -> string).
		// Zero-check on original type, then MarshalAmino to get repr.
		origZeroCheck := ctx.zeroCheckOriginal(accessor, finfo)
		if origZeroCheck != "" && !field.WriteEmpty {
			sb.WriteString(fmt.Sprintf("\tif %s {\n", origZeroCheck))
			sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			ctx.writeFieldValueMarshal(sb, "repr", fnum, finfo.ReprType, fopts, false, "\t\t")
			sb.WriteString("\t}\n")
		} else {
			sb.WriteString(fmt.Sprintf("\t{\n"))
			sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			ctx.writeFieldValueMarshal(sb, "repr", fnum, finfo.ReprType, fopts, field.WriteEmpty, "\t\t")
			sb.WriteString("\t}\n")
		}
		return nil
	}
	if finfo.IsAminoMarshaler && finfo.Type.Kind() == reflect.Struct {
		// Struct AminoMarshaler: MarshalAmino returns a repr, encode that.
		sb.WriteString(fmt.Sprintf("\t{\n"))
		sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		rinfo := finfo.ReprType
		if rinfo.Type.Kind() == reflect.Struct {
			sb.WriteString("\t\tvar buf bytes.Buffer\n")
			sb.WriteString("\t\tif err := repr.MarshalBinary2(cdc, &buf); err != nil {\n\t\t\treturn err\n\t\t}\n")
			ctx.writeBufAsField(sb, fnum, finfo.GetTyp3(fopts), field.WriteEmpty, "\t", "w")
		} else {
			ctx.writeFieldValueMarshal(sb, "repr", fnum, rinfo, fopts, field.WriteEmpty, "\t\t")
		}
		sb.WriteString("\t}\n")
		return nil
	}

	// Handle non-pointer fields: skip if default/zero value (unless WriteEmpty).
	if !field.WriteEmpty {
		zeroCheck := ctx.zeroCheck(accessor, finfo, fopts)
		if zeroCheck != "" {
			sb.WriteString(fmt.Sprintf("\tif %s {\n", zeroCheck))
			ctx.writeFieldValueMarshal(sb, accessor, fnum, finfo, fopts, false, "\t\t")
			sb.WriteString("\t}\n")
			return nil
		}
	}

	ctx.writeFieldValueMarshal(sb, accessor, fnum, finfo, fopts, field.WriteEmpty, "\t")
	return nil
}

func (ctx *P3Context2) writeFieldMarshalInline(sb *strings.Builder, field amino.FieldInfo, recv, indent, writerVar string) {
	// Simplified inline version for nested struct encoding.
	finfo := field.TypeInfo
	fname := field.Name
	fnum := field.BinFieldNum
	fopts := field.FieldOptions

	accessor := fmt.Sprintf("%s.%s", recv, fname)
	zeroCheck := ctx.zeroCheck(accessor, finfo, fopts)
	if zeroCheck != "" {
		sb.WriteString(fmt.Sprintf("%sif %s {\n", indent, zeroCheck))
		ctx.writeFieldValueMarshalTo(sb, accessor, fnum, finfo, fopts, false, indent+"\t", writerVar)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		ctx.writeFieldValueMarshalTo(sb, accessor, fnum, finfo, fopts, false, indent, writerVar)
	}
}

// writeFieldValueMarshal writes the field key + value encoding to default writer "w".
func (ctx *P3Context2) writeFieldValueMarshal(sb *strings.Builder, accessor string, fnum uint32, finfo *amino.TypeInfo, fopts amino.FieldOptions, writeEmpty bool, indent string) {
	ctx.writeFieldValueMarshalTo(sb, accessor, fnum, finfo, fopts, writeEmpty, indent, "w")
}

// writeFieldValueMarshalTo writes field key + value to a named writer.
func (ctx *P3Context2) writeFieldValueMarshalTo(sb *strings.Builder, accessor string, fnum uint32, finfo *amino.TypeInfo, fopts amino.FieldOptions, writeEmpty bool, indent, writerVar string) {
	typ3 := finfo.GetTyp3(fopts)
	rinfo := finfo.ReprType

	// For struct and ByteLength types we need a buffer pattern (write to buf, then length-prefix).
	switch {
	case rinfo.Type == reflect.TypeOf(time.Time{}):
		// Time: encode to buffer, then write field key + length-prefixed.
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tvar buf bytes.Buffer\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeTime(&buf, %s); err != nil {\n%s\t\treturn err\n%s\t}\n", indent, accessor, indent, indent))
		ctx.writeBufAsField(sb, fnum, typ3, writeEmpty, indent, writerVar)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tvar buf bytes.Buffer\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeDuration(&buf, %s); err != nil {\n%s\t\treturn err\n%s\t}\n", indent, accessor, indent, indent))
		ctx.writeBufAsField(sb, fnum, typ3, writeEmpty, indent, writerVar)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type.Kind() == reflect.Struct:
		// Nested struct: encode to buffer, then length-prefix.
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tvar buf bytes.Buffer\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tif err := %s.MarshalBinary2(cdc, &buf); err != nil {\n%s\t\treturn err\n%s\t}\n", indent, accessor, indent, indent))
		ctx.writeBufAsField(sb, fnum, typ3, writeEmpty, indent, writerVar)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type.Kind() == reflect.Interface:
		// Interface field: encode as google.protobuf.Any.
		ctx.writeInterfaceFieldMarshal(sb, accessor, fnum, indent, writerVar)

	case rinfo.Type.Kind() == reflect.String:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeFieldNumberAndTyp3(%s, %d, amino.Typ3ByteLength); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, fnum, indent, indent))
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeString(%s, string(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))

	case rinfo.Type.Kind() == reflect.Slice && rinfo.Type.Elem().Kind() == reflect.Uint8:
		// []byte
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeFieldNumberAndTyp3(%s, %d, amino.Typ3ByteLength); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, fnum, indent, indent))
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeByteSlice(%s, %s); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))

	case rinfo.Type.Kind() == reflect.Array && rinfo.Type.Elem().Kind() == reflect.Uint8:
		// [N]byte
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeFieldNumberAndTyp3(%s, %d, amino.Typ3ByteLength); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, fnum, indent, indent))
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeByteSlice(%s, %s[:]); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))

	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Packed list (non-byte elements with non-ByteLength typ3).
		// Encode all elements into a buffer, then write field key + length-prefixed.
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tvar buf bytes.Buffer\n", indent))
		einfo := finfo.Elem
		ert := rinfo.Type.Elem()
		eFopts := fopts
		if einfo != nil && einfo.ReprType.Type.Kind() == reflect.Uint8 {
			// List of (repr) bytes: encode as bytes.
			sb.WriteString(fmt.Sprintf("%s\tfor _, e := range %s {\n", indent, accessor))
			eAccessor := "e"
			if ert.Kind() == reflect.Ptr {
				sb.WriteString(fmt.Sprintf("%s\t\tvar de %s\n", indent, ctx.goTypeName(ert.Elem())))
				sb.WriteString(fmt.Sprintf("%s\t\tif e != nil {\n%s\t\t\tde = *e\n%s\t\t}\n", indent, indent, indent))
				eAccessor = "de"
			}
			if einfo.IsAminoMarshaler {
				sb.WriteString(fmt.Sprintf("%s\t\ter, err := %s.MarshalAmino()\n", indent, eAccessor))
				sb.WriteString(fmt.Sprintf("%s\t\tif err != nil {\n%s\t\t\treturn err\n%s\t\t}\n", indent, indent, indent))
				sb.WriteString(fmt.Sprintf("%s\t\tif err := amino.EncodeByte(&buf, uint8(er)); err != nil {\n%s\t\t\treturn err\n%s\t\t}\n", indent, indent, indent))
			} else {
				sb.WriteString(fmt.Sprintf("%s\t\tif err := amino.EncodeByte(&buf, uint8(%s)); err != nil {\n%s\t\t\treturn err\n%s\t\t}\n", indent, eAccessor, indent, indent))
			}
			sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
		} else if einfo != nil {
			sb.WriteString(fmt.Sprintf("%s\tfor _, e := range %s {\n", indent, accessor))
			if ert.Kind() == reflect.Ptr {
				sb.WriteString(fmt.Sprintf("%s\t\tde := e\n", indent))
				sb.WriteString(fmt.Sprintf("%s\t\tif de == nil {\n%s\t\t\tde = new(%s)\n%s\t\t}\n", indent, indent, ctx.goTypeName(ert.Elem()), indent))
				ctx.writePrimitiveEncode(sb, "(*de)", einfo, eFopts, indent+"\t\t", "&buf")
			} else {
				ctx.writePrimitiveEncode(sb, "e", einfo, eFopts, indent+"\t\t", "&buf")
			}
			sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
		}
		// For packed lists, always write: the outer len(slice)>0 guard ensures non-empty.
		// Don't apply the "skip if 0x00" check that writeFieldKeyAndBuf does,
		// because a slice like []int8{0} is non-empty and must be written.
		ctx.writeBufAsField(sb, fnum, typ3, true, indent, writerVar)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	default:
		// Primitive types.
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeFieldNumberAndTyp3(%s, %d, %s); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, fnum, typ3GoStr(typ3), indent, indent))
		ctx.writePrimitiveEncode(sb, accessor, rinfo, fopts, indent, writerVar)
	}
}

// writeBufAsField writes buf contents as a length-prefixed field.
// When writeEmpty is false, it skips the field if buf is empty or contains
// only a single 0x00 byte (matching amino's writeFieldIfNotEmpty behavior).
func (ctx *P3Context2) writeBufAsField(sb *strings.Builder, fnum uint32, typ3 amino.Typ3, writeEmpty bool, indent, writerVar string) {
	sb.WriteString(fmt.Sprintf("%s\tbz := buf.Bytes()\n", indent))
	if !writeEmpty {
		// Match amino's writeFieldIfNotEmpty: if encoded value is just 0x00, skip the field.
		sb.WriteString(fmt.Sprintf("%s\tif len(bz) == 0 || (len(bz) == 1 && bz[0] == 0x00) {\n", indent))
		sb.WriteString(fmt.Sprintf("%s\t\t// skip empty\n", indent))
		sb.WriteString(fmt.Sprintf("%s\t} else {\n", indent))
		sb.WriteString(fmt.Sprintf("%s\t\tif err := amino.EncodeFieldNumberAndTyp3(%s, %d, %s); err != nil {\n%s\t\t\treturn err\n%s\t\t}\n",
			indent, writerVar, fnum, typ3GoStr(typ3), indent, indent))
		sb.WriteString(fmt.Sprintf("%s\t\tif err := amino.EncodeByteSlice(%s, bz); err != nil {\n%s\t\t\treturn err\n%s\t\t}\n",
			indent, writerVar, indent, indent))
		sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
	} else {
		sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeFieldNumberAndTyp3(%s, %d, %s); err != nil {\n%s\t\treturn err\n%s\t}\n",
			indent, writerVar, fnum, typ3GoStr(typ3), indent, indent))
		sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeByteSlice(%s, bz); err != nil {\n%s\t\treturn err\n%s\t}\n",
			indent, writerVar, indent, indent))
	}
}

// === List / Repeated Field Encoding ===

func (ctx *P3Context2) writeUnpackedListMarshal(sb *strings.Builder, accessor string, finfo *amino.TypeInfo, fopts amino.FieldOptions) error {
	// Unpacked list: the slice is at the field level, encoded bare (repeated field entries).
	// This matches encodeReflectBinaryList with bare=true.
	ert := finfo.Type.Elem()
	einfo := finfo.Elem
	if einfo == nil {
		return fmt.Errorf("list type %v has nil Elem info", finfo.Type)
	}

	// beOptionByte: when element repr is uint8, amino encodes each element
	// as a raw byte rather than a varint (packed as a byte string).
	beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength || beOptionByte {
		// Packed form: write all elements into a buffer, then write field key + length prefix.
		sb.WriteString(fmt.Sprintf("\tif len(%s) > 0 {\n", accessor))
		sb.WriteString(fmt.Sprintf("\t\tif err := amino.EncodeFieldNumberAndTyp3(w, %d, amino.Typ3ByteLength); err != nil {\n\t\t\treturn err\n\t\t}\n", fopts.BinFieldNum))
		sb.WriteString("\t\tvar buf bytes.Buffer\n")
		sb.WriteString(fmt.Sprintf("\t\tfor _, elem := range %s {\n", accessor))
		if ert.Kind() == reflect.Ptr {
			sb.WriteString("\t\t\tif elem == nil {\n")
			sb.WriteString("\t\t\t\telem = new(" + ert.Elem().Name() + ")\n")
			sb.WriteString("\t\t\t}\n")
			ctx.writePrimitiveEncode(sb, "(*elem)", einfo, fopts, "\t\t\t", "&buf")
		} else {
			ctx.writePrimitiveEncode(sb, "elem", einfo, fopts, "\t\t\t", "&buf")
		}
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tif err := amino.EncodeByteSlice(w, buf.Bytes()); err != nil {\n\t\t\treturn err\n\t\t}\n")
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field key per element.
		ertIsPointer := ert.Kind() == reflect.Ptr
		writeImplicit := isListType(einfo.Type) &&
			einfo.Elem != nil &&
			einfo.Elem.ReprType.Type.Kind() != reflect.Uint8 &&
			einfo.Elem.ReprType.GetTyp3(fopts) != amino.Typ3ByteLength

		sb.WriteString(fmt.Sprintf("\tfor _, elem := range %s {\n", accessor))
		sb.WriteString(fmt.Sprintf("\t\tif err := amino.EncodeFieldNumberAndTyp3(w, %d, amino.Typ3ByteLength); err != nil {\n\t\t\treturn err\n\t\t}\n", fopts.BinFieldNum))

		if ertIsPointer {
			ertIsStruct := einfo.ReprType.Type.Kind() == reflect.Struct
			sb.WriteString("\t\tif elem == nil {\n")
			if ertIsStruct && !fopts.NilElements {
				// Match amino: nil struct pointers in lists error unless nil_elements tag is set.
				sb.WriteString("\t\t\treturn errors.New(\"nil struct pointers in lists not supported unless nil_elements field tag is also set\")\n")
			} else {
				sb.WriteString("\t\t\tif err := amino.EncodeByte(w, 0x00); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
			}
			sb.WriteString("\t\t} else {\n")
		}

		elemAccessor := "elem"
		extraIndent := "\t\t"
		if ertIsPointer {
			elemAccessor = "(*elem)"
			extraIndent = "\t\t\t"
		}

		if einfo.ReprType.Type.Kind() == reflect.Interface {
			// Interface element: encode via MarshalAny.
			sb.WriteString(fmt.Sprintf("%sif %s != nil {\n", extraIndent, elemAccessor))
			sb.WriteString(fmt.Sprintf("%s\tanyBz, err := cdc.MarshalAny(%s)\n", extraIndent, elemAccessor))
			sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn err\n%s\t}\n", extraIndent, extraIndent, extraIndent))
			sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeByteSlice(w, anyBz); err != nil {\n%s\t\treturn err\n%s\t}\n",
				extraIndent, extraIndent, extraIndent))
			sb.WriteString(fmt.Sprintf("%s} else {\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeByte(w, 0x00); err != nil {\n%s\t\treturn err\n%s\t}\n",
				extraIndent, extraIndent, extraIndent))
			sb.WriteString(fmt.Sprintf("%s}\n", extraIndent))
		} else if writeImplicit {
			// Nested list: wrap in implicit struct.
			// If the inner list is empty, the implicit struct is empty (no inner field).
			sb.WriteString(fmt.Sprintf("%svar ebuf bytes.Buffer\n", extraIndent))
			// Encode element list to ebuf.
			ctx.writeListEncodeToBuf(sb, elemAccessor, einfo, fopts, extraIndent, "&ebuf")
			sb.WriteString(fmt.Sprintf("%svar ibuf bytes.Buffer\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%sif ebuf.Len() > 0 {\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeFieldNumberAndTyp3(&ibuf, 1, amino.Typ3ByteLength); err != nil {\n%s\t\treturn err\n%s\t}\n",
				extraIndent, extraIndent, extraIndent))
			sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeByteSlice(&ibuf, ebuf.Bytes()); err != nil {\n%s\t\treturn err\n%s\t}\n",
				extraIndent, extraIndent, extraIndent))
			sb.WriteString(fmt.Sprintf("%s}\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeByteSlice(w, ibuf.Bytes()); err != nil {\n%s\treturn err\n%s}\n",
				extraIndent, extraIndent, extraIndent))
		} else if einfo.ReprType.Type.Kind() == reflect.Struct ||
			einfo.ReprType.Type == reflect.TypeOf(time.Duration(0)) ||
			(isListType(einfo.ReprType.Type) && einfo.ReprType.Type.Elem().Kind() != reflect.Uint8) {
			// Struct/Duration/nested-list element: encode to buf, then length-prefix.
			// (cdc.Marshal and MarshalBinary2 produce bare content for these)
			sb.WriteString(fmt.Sprintf("%svar buf bytes.Buffer\n", extraIndent))
			ctx.writeElementEncodeToBuf(sb, elemAccessor, einfo, fopts, extraIndent, "&buf")
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeByteSlice(w, buf.Bytes()); err != nil {\n%s\treturn err\n%s}\n",
				extraIndent, extraIndent, extraIndent))
		} else {
			// Non-struct ByteLength element (string, []byte):
			// encode directly to w (encode functions include own length prefix).
			ctx.writeElementEncodeToBuf(sb, elemAccessor, einfo, fopts, extraIndent, "w")
		}

		if ertIsPointer {
			sb.WriteString("\t\t}\n")
		}
		sb.WriteString("\t}\n")
	}
	return nil
}

func (ctx *P3Context2) writeListEncodeToBuf(sb *strings.Builder, accessor string, info *amino.TypeInfo, fopts amino.FieldOptions, indent, writerVar string) {
	// Encode list elements packed into the writer.
	einfo := info.Elem
	ert := info.Type.Elem()
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength {
		sb.WriteString(fmt.Sprintf("%sfor _, e := range %s {\n", indent, accessor))
		if ert.Kind() == reflect.Ptr {
			sb.WriteString(fmt.Sprintf("%s\tif e == nil {\n%s\t\te = new(%s)\n%s\t}\n", indent, indent, ert.Elem().Name(), indent))
			ctx.writePrimitiveEncode(sb, "(*e)", einfo, fopts, indent+"\t", writerVar)
		} else {
			ctx.writePrimitiveEncode(sb, "e", einfo, fopts, indent+"\t", writerVar)
		}
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		sb.WriteString(fmt.Sprintf("%sfor _, e := range %s {\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeFieldNumberAndTyp3(%s, 1, amino.Typ3ByteLength); err != nil {\n%s\t\treturn err\n%s\t}\n",
			indent, writerVar, indent, indent))
		sb.WriteString(fmt.Sprintf("%s\tvar elbuf bytes.Buffer\n", indent))
		ctx.writeElementEncodeToBuf(sb, "e", einfo, fopts, indent+"\t", "&elbuf")
		sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeByteSlice(%s, elbuf.Bytes()); err != nil {\n%s\t\treturn err\n%s\t}\n",
			indent, writerVar, indent, indent))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	}
}

func (ctx *P3Context2) writeElementEncodeToBuf(sb *strings.Builder, accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions, indent, writerVar string) {
	rinfo := einfo.ReprType
	switch {
	case rinfo.Type == reflect.TypeOf(time.Time{}):
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeTime(%s, %s); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeDuration(%s, %s); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case rinfo.Type.Kind() == reflect.Struct:
		sb.WriteString(fmt.Sprintf("%sif err := %s.MarshalBinary2(cdc, %s); err != nil {\n%s\treturn err\n%s}\n",
			indent, accessor, writerVar, indent, indent))
	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Nested list element (e.g., [2]string, [][]byte): fall back to cdc.Marshal.
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tbz, err := cdc.Marshal(%s)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn err\n%s\t}\n", indent, indent, indent))
		// Use a writer variable that works with both "w" and "&buf" style names.
		sb.WriteString(fmt.Sprintf("%s\tif _, werr := (%s).Write(bz); werr != nil {\n%s\t\treturn werr\n%s\t}\n", indent, writerVar, indent, indent))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	default:
		ctx.writePrimitiveEncode(sb, accessor, einfo, fopts, indent, writerVar)
	}
}

func (ctx *P3Context2) writeInterfaceFieldMarshal(sb *strings.Builder, accessor string, fnum uint32, indent, writerVar string) {
	sb.WriteString(fmt.Sprintf("%sif %s != nil {\n", indent, accessor))
	// MarshalAny encodes as google.protobuf.Any (type_url + value) bare.
	sb.WriteString(fmt.Sprintf("%s\tanyBz, err := cdc.MarshalAny(%s)\n", indent, accessor))
	sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn err\n%s\t}\n", indent, indent, indent))
	// Write outer field key + length-prefix.
	sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeFieldNumberAndTyp3(%s, %d, amino.Typ3ByteLength); err != nil {\n%s\t\treturn err\n%s\t}\n",
		indent, writerVar, fnum, indent, indent))
	sb.WriteString(fmt.Sprintf("%s\tif err := amino.EncodeByteSlice(%s, anyBz); err != nil {\n%s\t\treturn err\n%s\t}\n",
		indent, writerVar, indent, indent))
	sb.WriteString(fmt.Sprintf("%s}\n", indent))
}

// === Primitive Encoding ===

func (ctx *P3Context2) writePrimitiveEncode(sb *strings.Builder, accessor string, info *amino.TypeInfo, fopts amino.FieldOptions, indent, writerVar string) {
	rinfo := info.ReprType
	rt := rinfo.Type
	kind := rt.Kind()

	switch kind {
	case reflect.Bool:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeBool(%s, bool(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case reflect.Int8:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeVarint(%s, int64(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case reflect.Int16:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeVarint(%s, int64(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case reflect.Int32:
		if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeInt32(%s, int32(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeVarint(%s, int64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		}
	case reflect.Int64:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeInt64(%s, int64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeVarint(%s, int64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		}
	case reflect.Int:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeInt64(%s, int64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeInt32(%s, int32(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeVarint(%s, int64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		}
	case reflect.Uint8:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUvarint(%s, uint64(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case reflect.Uint16:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUvarint(%s, uint64(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case reflect.Uint32:
		if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUint32(%s, uint32(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUvarint(%s, uint64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		}
	case reflect.Uint64:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUint64(%s, uint64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUvarint(%s, uint64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		}
	case reflect.Uint:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUint64(%s, uint64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUint32(%s, uint32(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeUvarint(%s, uint64(%s)); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		}
	case reflect.Float32:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeFloat32(%s, float32(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case reflect.Float64:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeFloat64(%s, float64(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case reflect.String:
		sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeString(%s, string(%s)); err != nil {\n%s\treturn err\n%s}\n",
			indent, writerVar, accessor, indent, indent))
	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeByteSlice(%s, %s); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive slice element kind %v\n", indent, rt.Elem().Kind()))
		}
	case reflect.Array:
		if rt.Elem().Kind() == reflect.Uint8 {
			sb.WriteString(fmt.Sprintf("%sif err := amino.EncodeByteSlice(%s, %s[:]); err != nil {\n%s\treturn err\n%s}\n",
				indent, writerVar, accessor, indent, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive array element kind %v\n", indent, rt.Elem().Kind()))
		}
	default:
		sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive kind %v\n", indent, kind))
	}
}

// writeValueEncode writes a single value encode (no field key) for packed list elements.
func (ctx *P3Context2) writeValueEncode(sb *strings.Builder, indent, accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions) {
	ctx.writePrimitiveEncode(sb, accessor, einfo, fopts, indent, "&buf")
}

// === Helpers ===

// zeroCheck returns a Go expression that is true when the value is NOT the zero value.
// Returns "" if no zero check should be applied (e.g. for structs).
func (ctx *P3Context2) zeroCheck(accessor string, info *amino.TypeInfo, fopts amino.FieldOptions) string {
	rinfo := info.ReprType
	rt := rinfo.Type

	// time.Duration is treated as a struct (not skipped on zero).
	if rt == reflect.TypeOf(time.Duration(0)) {
		return ""
	}

	switch rt.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%s", accessor) // bool is truthy check
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%s != 0", accessor)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%s != 0", accessor)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%s != 0", accessor)
	case reflect.String:
		return fmt.Sprintf("%s != \"\"", accessor)
	case reflect.Slice:
		return fmt.Sprintf("len(%s) != 0", accessor)
	case reflect.Struct:
		return "" // structs are never "default" per isNonstructDefaultValue
	case reflect.Interface:
		return fmt.Sprintf("%s != nil", accessor)
	default:
		return ""
	}
}

// zeroCheckOriginal generates a zero-value check based on the original type (not repr).
// Used for AminoMarshaler fields where the repr type differs from the original.
func (ctx *P3Context2) zeroCheckOriginal(accessor string, info *amino.TypeInfo) string {
	rt := info.Type
	switch rt.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%s", accessor)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%s != 0", accessor)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%s != 0", accessor)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%s != 0", accessor)
	case reflect.String:
		return fmt.Sprintf("%s != \"\"", accessor)
	case reflect.Slice:
		return fmt.Sprintf("len(%s) != 0", accessor)
	case reflect.Array:
		// For byte arrays, check if not all zeros.
		if rt.Elem().Kind() == reflect.Uint8 {
			return fmt.Sprintf("%s != [%d]byte{}", accessor, rt.Len())
		}
		return ""
	case reflect.Struct:
		return ""
	case reflect.Interface:
		return fmt.Sprintf("%s != nil", accessor)
	default:
		return ""
	}
}

func typ3GoStr(t amino.Typ3) string {
	switch t {
	case amino.Typ3Varint:
		return "amino.Typ3Varint"
	case amino.Typ38Byte:
		return "amino.Typ38Byte"
	case amino.Typ3ByteLength:
		return "amino.Typ3ByteLength"
	case amino.Typ34Byte:
		return "amino.Typ34Byte"
	default:
		return fmt.Sprintf("amino.Typ3(%d)", t)
	}
}

func isListType(rt reflect.Type) bool {
	return rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array
}

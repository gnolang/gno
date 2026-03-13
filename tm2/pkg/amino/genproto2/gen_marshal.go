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

	sb.WriteString(fmt.Sprintf("func (goo %s) MarshalBinary2(cdc *amino.Codec, buf []byte, offset int) (int, error) {\n", tname))
	sb.WriteString("\tvar err error\n")

	// Handle AminoMarshaler: convert to repr, then marshal repr.
	if info.IsAminoMarshaler {
		sb.WriteString("\trepr, err := goo.MarshalAmino()\n")
		sb.WriteString("\tif err != nil {\n\t\treturn offset, err\n\t}\n")
		rinfo := info.ReprType
		if err := ctx.writeReprMarshal(sb, rinfo); err != nil {
			return err
		}
		sb.WriteString("\treturn offset, err\n")
		sb.WriteString("}\n\n")
		return nil
	}

	if err := ctx.writeStructMarshalBody(sb, info, "goo"); err != nil {
		return err
	}

	sb.WriteString("\treturn offset, err\n")
	sb.WriteString("}\n\n")
	return nil
}

// === AminoMarshaler Repr Handling ===

func (ctx *P3Context2) writeReprMarshal(sb *strings.Builder, rinfo *amino.TypeInfo) error {
	rt := rinfo.Type
	fopts := amino.FieldOptions{}

	switch {
	case rt.Kind() == reflect.Struct:
		return ctx.writeStructMarshalBody(sb, rinfo, "repr")

	case isListType(rt):
		if !rinfo.IsStructOrUnpacked(fopts) {
			return ctx.writePackedSliceReprMarshal(sb, rinfo)
		}
		return ctx.writeSliceReprMarshal(sb, rinfo)

	default:
		// Primitive repr wrapped in implicit struct field 1.
		typ3 := rinfo.GetTyp3(fopts)
		sb.WriteString("\t{\n")
		sb.WriteString("\t\tbefore := offset\n")
		ctx.writePrimitiveEncode(sb, "repr", rinfo, fopts, "\t\t")
		// Match writeFieldIfNotEmpty: skip if value encoded to nothing or single zero byte.
		sb.WriteString("\t\tvalueLen := before - offset\n")
		sb.WriteString("\t\tif valueLen > 0 {\n")
		sb.WriteString(fmt.Sprintf("\t\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, %s)\n", typ3GoStr(typ3)))
		sb.WriteString("\t\t} else {\n")
		sb.WriteString("\t\t\toffset = before\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
	}

	return nil
}

func (ctx *P3Context2) writePackedSliceReprMarshal(sb *strings.Builder, info *amino.TypeInfo) error {
	einfo := info.Elem
	if einfo == nil {
		return fmt.Errorf("slice type %v has nil Elem info", info.Type)
	}

	fopts := amino.FieldOptions{}
	beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8
	sb.WriteString("\tif len(repr) > 0 {\n")
	sb.WriteString("\t\tbefore := offset\n")
	sb.WriteString("\t\tfor i := len(repr) - 1; i >= 0; i-- {\n")
	sb.WriteString("\t\t\telem := repr[i]\n")
	if beOptionByte {
		if einfo.IsAminoMarshaler {
			sb.WriteString("\t\t\telemRepr, err := elem.MarshalAmino()\n")
			sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn offset, err\n\t\t\t}\n")
			sb.WriteString("\t\t\toffset = amino.PrependByte(buf, offset, uint8(elemRepr))\n")
		} else {
			sb.WriteString("\t\t\toffset = amino.PrependByte(buf, offset, uint8(elem))\n")
		}
	} else if einfo.IsAminoMarshaler {
		sb.WriteString("\t\t\telemRepr, err := elem.MarshalAmino()\n")
		sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn offset, err\n\t\t\t}\n")
		ctx.writePrimitiveEncode(sb, "elemRepr", einfo.ReprType, fopts, "\t\t\t")
	} else {
		ctx.writePrimitiveEncode(sb, "elem", einfo, fopts, "\t\t\t")
	}
	sb.WriteString("\t\t}\n")
	// Wrap in field 1 key + length prefix.
	sb.WriteString("\t\tdataLen := before - offset\n")
	sb.WriteString("\t\toffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n")
	sb.WriteString("\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n")
	sb.WriteString("\t}\n")
	return nil
}

func (ctx *P3Context2) writeSliceReprMarshal(sb *strings.Builder, info *amino.TypeInfo) error {
	einfo := info.Elem
	if einfo == nil {
		return fmt.Errorf("slice type %v has nil Elem info", info.Type)
	}

	fopts := amino.FieldOptions{}
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength {
		// Packed form: single length-prefixed block.
		sb.WriteString("\tif len(repr) > 0 {\n")
		sb.WriteString("\t\tbefore := offset\n")
		sb.WriteString("\t\tfor i := len(repr) - 1; i >= 0; i-- {\n")
		sb.WriteString("\t\t\telem := repr[i]\n")
		if einfo.IsAminoMarshaler {
			sb.WriteString("\t\t\telemRepr, err := elem.MarshalAmino()\n")
			sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn offset, err\n\t\t\t}\n")
			ctx.writePrimitiveEncode(sb, "elemRepr", einfo.ReprType, fopts, "\t\t\t")
		} else {
			ctx.writePrimitiveEncode(sb, "elem", einfo, fopts, "\t\t\t")
		}
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tdataLen := before - offset\n")
		sb.WriteString("\t\toffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n")
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field entries with field number 1.
		sb.WriteString("\tfor i := len(repr) - 1; i >= 0; i-- {\n")
		sb.WriteString("\t\telem := repr[i]\n")
		if einfo.Type.Kind() == reflect.Struct {
			sb.WriteString("\t\tbefore := offset\n")
			if einfo.IsAminoMarshaler {
				sb.WriteString("\t\telemRepr, err := elem.MarshalAmino()\n")
				sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
				ctx.writeStructMarshalBodyInline(sb, einfo.ReprType, "elemRepr", "\t\t")
			} else {
				sb.WriteString("\t\toffset, err = elem.MarshalBinary2(cdc, buf, offset)\n")
				sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
			}
			sb.WriteString("\t\tdataLen := before - offset\n")
			sb.WriteString("\t\toffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n")
			sb.WriteString("\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n")
		} else {
			ctx.writePrimitiveEncode(sb, "elem", einfo, fopts, "\t\t")
			sb.WriteString("\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n")
		}
		sb.WriteString("\t}\n")
	}
	return nil
}

// === Struct Fields ===

func (ctx *P3Context2) writeStructMarshalBody(sb *strings.Builder, info *amino.TypeInfo, recv string) error {
	// Write fields in REVERSE order for backward encoding.
	for i := len(info.Fields) - 1; i >= 0; i-- {
		if err := ctx.writeFieldMarshal(sb, info.Fields[i], recv); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *P3Context2) writeStructMarshalBodyInline(sb *strings.Builder, info *amino.TypeInfo, recv, indent string) {
	for i := len(info.Fields) - 1; i >= 0; i-- {
		ctx.writeFieldMarshalInline(sb, info.Fields[i], recv, indent)
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
		origZeroCheck := ctx.zeroCheckOriginal(accessor, finfo)
		if origZeroCheck != "" && !field.WriteEmpty {
			sb.WriteString(fmt.Sprintf("\tif %s {\n", origZeroCheck))
			sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
			ctx.writeFieldValueMarshal(sb, "repr", fnum, finfo.ReprType, fopts, false, "\t\t")
			sb.WriteString("\t}\n")
		} else {
			sb.WriteString(fmt.Sprintf("\t{\n"))
			sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
			ctx.writeFieldValueMarshal(sb, "repr", fnum, finfo.ReprType, fopts, field.WriteEmpty, "\t\t")
			sb.WriteString("\t}\n")
		}
		return nil
	}
	if finfo.IsAminoMarshaler && finfo.Type.Kind() == reflect.Struct {
		sb.WriteString(fmt.Sprintf("\t{\n"))
		sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
		rinfo := finfo.ReprType
		if rinfo.Type.Kind() == reflect.Struct {
			sb.WriteString("\t\tbefore := offset\n")
			sb.WriteString("\t\toffset, err = repr.MarshalBinary2(cdc, buf, offset)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
			ctx.writeLengthPrefixedField(sb, fnum, finfo.GetTyp3(fopts), field.WriteEmpty, "\t\t")
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

func (ctx *P3Context2) writeFieldMarshalInline(sb *strings.Builder, field amino.FieldInfo, recv, indent string) {
	finfo := field.TypeInfo
	fname := field.Name
	fnum := field.BinFieldNum
	fopts := field.FieldOptions

	accessor := fmt.Sprintf("%s.%s", recv, fname)
	zeroCheck := ctx.zeroCheck(accessor, finfo, fopts)
	if zeroCheck != "" {
		sb.WriteString(fmt.Sprintf("%sif %s {\n", indent, zeroCheck))
		ctx.writeFieldValueMarshal(sb, accessor, fnum, finfo, fopts, false, indent+"\t")
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		ctx.writeFieldValueMarshal(sb, accessor, fnum, finfo, fopts, false, indent)
	}
}

// writeFieldValueMarshal writes field value + field key backward.
func (ctx *P3Context2) writeFieldValueMarshal(sb *strings.Builder, accessor string, fnum uint32, finfo *amino.TypeInfo, fopts amino.FieldOptions, writeEmpty bool, indent string) {
	typ3 := finfo.GetTyp3(fopts)
	rinfo := finfo.ReprType

	switch {
	case rinfo.Type == reflect.TypeOf(time.Time{}):
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tbefore := offset\n", indent))
		sb.WriteString(fmt.Sprintf("%s\toffset, err = amino.PrependTime(buf, offset, %s)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent))
		ctx.writeLengthPrefixedField(sb, fnum, typ3, writeEmpty, indent+"\t")
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tbefore := offset\n", indent))
		sb.WriteString(fmt.Sprintf("%s\toffset, err = amino.PrependDuration(buf, offset, %s)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent))
		ctx.writeLengthPrefixedField(sb, fnum, typ3, writeEmpty, indent+"\t")
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type.Kind() == reflect.Struct:
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tbefore := offset\n", indent))
		sb.WriteString(fmt.Sprintf("%s\toffset, err = %s.MarshalBinary2(cdc, buf, offset)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent))
		ctx.writeLengthPrefixedField(sb, fnum, typ3, writeEmpty, indent+"\t")
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type.Kind() == reflect.Interface:
		ctx.writeInterfaceFieldMarshal(sb, accessor, fnum, indent)

	case rinfo.Type.Kind() == reflect.String:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependString(buf, offset, string(%s))\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", indent, fnum))

	case rinfo.Type.Kind() == reflect.Slice && rinfo.Type.Elem().Kind() == reflect.Uint8:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependByteSlice(buf, offset, %s)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", indent, fnum))

	case rinfo.Type.Kind() == reflect.Array && rinfo.Type.Elem().Kind() == reflect.Uint8:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependByteSlice(buf, offset, %s[:])\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", indent, fnum))

	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Packed list (non-byte elements).
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tbefore := offset\n", indent))
		einfo := finfo.Elem
		ert := rinfo.Type.Elem()
		eFopts := fopts
		if einfo != nil && einfo.ReprType.Type.Kind() == reflect.Uint8 {
			// List of (repr) bytes.
			sb.WriteString(fmt.Sprintf("%s\tfor i := len(%s) - 1; i >= 0; i-- {\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%s\t\te := %s[i]\n", indent, accessor))
			eAccessor := "e"
			if ert.Kind() == reflect.Ptr {
				sb.WriteString(fmt.Sprintf("%s\t\tvar de %s\n", indent, ctx.goTypeName(ert.Elem())))
				sb.WriteString(fmt.Sprintf("%s\t\tif e != nil {\n%s\t\t\tde = *e\n%s\t\t}\n", indent, indent, indent))
				eAccessor = "de"
			}
			if einfo.IsAminoMarshaler {
				sb.WriteString(fmt.Sprintf("%s\t\ter, err := %s.MarshalAmino()\n", indent, eAccessor))
				sb.WriteString(fmt.Sprintf("%s\t\tif err != nil {\n%s\t\t\treturn offset, err\n%s\t\t}\n", indent, indent, indent))
				sb.WriteString(fmt.Sprintf("%s\t\toffset = amino.PrependByte(buf, offset, uint8(er))\n", indent))
			} else {
				sb.WriteString(fmt.Sprintf("%s\t\toffset = amino.PrependByte(buf, offset, uint8(%s))\n", indent, eAccessor))
			}
			sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
		} else if einfo != nil {
			sb.WriteString(fmt.Sprintf("%s\tfor i := len(%s) - 1; i >= 0; i-- {\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%s\t\te := %s[i]\n", indent, accessor))
			if ert.Kind() == reflect.Ptr {
				sb.WriteString(fmt.Sprintf("%s\t\tif e == nil {\n%s\t\t\te = new(%s)\n%s\t\t}\n", indent, indent, ctx.goTypeName(ert.Elem()), indent))
				ctx.writePrimitiveEncode(sb, "(*e)", einfo, eFopts, indent+"\t\t")
			} else {
				ctx.writePrimitiveEncode(sb, "e", einfo, eFopts, indent+"\t\t")
			}
			sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
		}
		// For packed lists, always write (outer len check ensures non-empty).
		ctx.writeLengthPrefixedField(sb, fnum, typ3, true, indent+"\t")
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	default:
		// Primitive types: value first, then field key.
		ctx.writePrimitiveEncode(sb, accessor, rinfo, fopts, indent)
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, %s)\n",
			indent, fnum, typ3GoStr(typ3)))
	}
}

// writeLengthPrefixedField writes the length prefix + field key after data has been
// written backward. Assumes `before` and `offset` are in scope.
func (ctx *P3Context2) writeLengthPrefixedField(sb *strings.Builder, fnum uint32, typ3 amino.Typ3, writeEmpty bool, indent string) {
	sb.WriteString(fmt.Sprintf("%sdataLen := before - offset\n", indent))
	if !writeEmpty {
		sb.WriteString(fmt.Sprintf("%sif dataLen > 0 {\n", indent))
		sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n", indent))
		sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, %s)\n", indent, fnum, typ3GoStr(typ3)))
		sb.WriteString(fmt.Sprintf("%s} else {\n", indent))
		sb.WriteString(fmt.Sprintf("%s\toffset = before\n", indent))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n", indent))
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, %s)\n", indent, fnum, typ3GoStr(typ3)))
	}
}

// === List / Repeated Field Encoding ===

func (ctx *P3Context2) writeUnpackedListMarshal(sb *strings.Builder, accessor string, finfo *amino.TypeInfo, fopts amino.FieldOptions) error {
	ert := finfo.Type.Elem()
	einfo := finfo.Elem
	if einfo == nil {
		return fmt.Errorf("list type %v has nil Elem info", finfo.Type)
	}

	beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength || beOptionByte {
		// Packed form: field key + length + packed content.
		sb.WriteString(fmt.Sprintf("\tif len(%s) > 0 {\n", accessor))
		sb.WriteString("\t\tbefore := offset\n")
		sb.WriteString(fmt.Sprintf("\t\tfor i := len(%s) - 1; i >= 0; i-- {\n", accessor))
		sb.WriteString(fmt.Sprintf("\t\t\telem := %s[i]\n", accessor))
		if ert.Kind() == reflect.Ptr {
			sb.WriteString("\t\t\tif elem == nil {\n")
			sb.WriteString("\t\t\t\telem = new(" + ert.Elem().Name() + ")\n")
			sb.WriteString("\t\t\t}\n")
			ctx.writePrimitiveEncode(sb, "(*elem)", einfo, fopts, "\t\t\t")
		} else {
			ctx.writePrimitiveEncode(sb, "elem", einfo, fopts, "\t\t\t")
		}
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tdataLen := before - offset\n")
		sb.WriteString("\t\toffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n")
		sb.WriteString(fmt.Sprintf("\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", fopts.BinFieldNum))
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field key per element.
		ertIsPointer := ert.Kind() == reflect.Ptr
		writeImplicit := isListType(einfo.Type) &&
			einfo.Elem != nil &&
			einfo.Elem.ReprType.Type.Kind() != reflect.Uint8 &&
			einfo.Elem.ReprType.GetTyp3(fopts) != amino.Typ3ByteLength

		sb.WriteString(fmt.Sprintf("\tfor i := len(%s) - 1; i >= 0; i-- {\n", accessor))
		sb.WriteString(fmt.Sprintf("\t\telem := %s[i]\n", accessor))

		elemAccessor := "elem"
		extraIndent := "\t\t"

		if ertIsPointer {
			ertIsStruct := einfo.ReprType.Type.Kind() == reflect.Struct
			sb.WriteString("\t\tif elem == nil {\n")
			if ertIsStruct && !fopts.NilElements {
				sb.WriteString("\t\t\treturn offset, errors.New(\"nil struct pointers in lists not supported unless nil_elements field tag is also set\")\n")
			} else {
				sb.WriteString("\t\t\toffset = amino.PrependByte(buf, offset, 0x00)\n")
				sb.WriteString(fmt.Sprintf("\t\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", fopts.BinFieldNum))
			}
			sb.WriteString("\t\t} else {\n")
			elemAccessor = "(*elem)"
			extraIndent = "\t\t\t"
		}

		if einfo.ReprType.Type.Kind() == reflect.Interface {
			// Interface element: encode via MarshalAny.
			sb.WriteString(fmt.Sprintf("%sif %s != nil {\n", extraIndent, elemAccessor))
			sb.WriteString(fmt.Sprintf("%s\tanyBz, err := cdc.MarshalAny(%s)\n", extraIndent, elemAccessor))
			sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", extraIndent, extraIndent, extraIndent))
			sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependByteSlice(buf, offset, anyBz)\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s} else {\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependByte(buf, offset, 0x00)\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s}\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", extraIndent, fopts.BinFieldNum))
		} else if writeImplicit {
			// Nested list: wrap in implicit struct.
			sb.WriteString(fmt.Sprintf("%s{\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\timplicitBefore := offset\n", extraIndent))
			// Encode inner list elements.
			ctx.writeListEncode(sb, elemAccessor, einfo, fopts, extraIndent+"\t")
			sb.WriteString(fmt.Sprintf("%s\tinnerLen := implicitBefore - offset\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\tif innerLen > 0 {\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\t\toffset = amino.PrependUvarint(buf, offset, uint64(innerLen))\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\t}\n", extraIndent))
			// Outer: compute implicit struct size and write length prefix + field key.
			sb.WriteString(fmt.Sprintf("%s\tissLen := implicitBefore - offset\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependUvarint(buf, offset, uint64(issLen))\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s}\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", extraIndent, fopts.BinFieldNum))
		} else if einfo.ReprType.Type.Kind() == reflect.Struct ||
			einfo.ReprType.Type == reflect.TypeOf(time.Duration(0)) ||
			(isListType(einfo.ReprType.Type) && einfo.ReprType.Type.Elem().Kind() != reflect.Uint8) {
			// Struct/Duration/nested-list element: encode backward, then length-prefix.
			sb.WriteString(fmt.Sprintf("%sbefore := offset\n", extraIndent))
			ctx.writeElementEncode(sb, elemAccessor, einfo, fopts, extraIndent)
			sb.WriteString(fmt.Sprintf("%sdataLen := before - offset\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", extraIndent, fopts.BinFieldNum))
		} else {
			// Non-struct ByteLength element (string, []byte): includes own length prefix.
			ctx.writeElementEncode(sb, elemAccessor, einfo, fopts, extraIndent)
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", extraIndent, fopts.BinFieldNum))
		}

		if ertIsPointer {
			sb.WriteString("\t\t}\n")
		}
		sb.WriteString("\t}\n")
	}
	return nil
}

func (ctx *P3Context2) writeListEncode(sb *strings.Builder, accessor string, info *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	einfo := info.Elem
	ert := info.Type.Elem()
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength {
		sb.WriteString(fmt.Sprintf("%sfor i := len(%s) - 1; i >= 0; i-- {\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\te := %s[i]\n", indent, accessor))
		if ert.Kind() == reflect.Ptr {
			sb.WriteString(fmt.Sprintf("%s\tif e == nil {\n%s\t\te = new(%s)\n%s\t}\n", indent, indent, ert.Elem().Name(), indent))
			ctx.writePrimitiveEncode(sb, "(*e)", einfo, fopts, indent+"\t")
		} else {
			ctx.writePrimitiveEncode(sb, "e", einfo, fopts, indent+"\t")
		}
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	} else {
		sb.WriteString(fmt.Sprintf("%sfor i := len(%s) - 1; i >= 0; i-- {\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\te := %s[i]\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\telbefore := offset\n", indent))
		ctx.writeElementEncode(sb, "e", einfo, fopts, indent+"\t")
		sb.WriteString(fmt.Sprintf("%s\telLen := elbefore - offset\n", indent))
		sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependUvarint(buf, offset, uint64(elLen))\n", indent))
		sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n", indent))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	}
}

func (ctx *P3Context2) writeElementEncode(sb *strings.Builder, accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	rinfo := einfo.ReprType
	switch {
	case rinfo.Type == reflect.TypeOf(time.Time{}):
		sb.WriteString(fmt.Sprintf("%soffset, err = amino.PrependTime(buf, offset, %s)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn offset, err\n%s}\n", indent, indent, indent))
	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		sb.WriteString(fmt.Sprintf("%soffset, err = amino.PrependDuration(buf, offset, %s)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn offset, err\n%s}\n", indent, indent, indent))
	case rinfo.Type.Kind() == reflect.Struct:
		sb.WriteString(fmt.Sprintf("%soffset, err = %s.MarshalBinary2(cdc, buf, offset)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn offset, err\n%s}\n", indent, indent, indent))
	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Nested list element: encode as implicit struct content inline.
		// The caller handles the outer length prefix and field key.
		innerEinfo := einfo.Elem
		if innerEinfo == nil {
			// Fallback for unexpected cases.
			sb.WriteString(fmt.Sprintf("%s{\n", indent))
			sb.WriteString(fmt.Sprintf("%s\tbz, merr := cdc.Marshal(%s)\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%s\tif merr != nil {\n%s\t\treturn offset, merr\n%s\t}\n", indent, indent, indent))
			sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependBytes(buf, offset, bz)\n", indent))
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
			return
		}
		innerTyp3 := innerEinfo.GetTyp3(fopts)
		innerRinfo := innerEinfo.ReprType
		if innerTyp3 != amino.Typ3ByteLength {
			// Packed inner elements: single field 1 + length prefix.
			sb.WriteString(fmt.Sprintf("%s{\n", indent))
			sb.WriteString(fmt.Sprintf("%s\tpkBefore := offset\n", indent))
			sb.WriteString(fmt.Sprintf("%s\tfor ii := len(%s) - 1; ii >= 0; ii-- {\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%s\t\tie := %s[ii]\n", indent, accessor))
			ctx.writePrimitiveEncode(sb, "ie", innerEinfo, fopts, indent+"\t\t")
			sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
			sb.WriteString(fmt.Sprintf("%s\tpkLen := pkBefore - offset\n", indent))
			sb.WriteString(fmt.Sprintf("%s\tif pkLen > 0 {\n", indent))
			sb.WriteString(fmt.Sprintf("%s\t\toffset = amino.PrependUvarint(buf, offset, uint64(pkLen))\n", indent))
			sb.WriteString(fmt.Sprintf("%s\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n", indent))
			sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
		} else {
			// ByteLength inner elements: repeated field 1 entries.
			sb.WriteString(fmt.Sprintf("%sfor ii := len(%s) - 1; ii >= 0; ii-- {\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%s\tie := %s[ii]\n", indent, accessor))
			switch {
			case innerRinfo.Type == reflect.TypeOf(time.Time{}):
				sb.WriteString(fmt.Sprintf("%s\tieBefore := offset\n", indent))
				sb.WriteString(fmt.Sprintf("%s\toffset, err = amino.PrependTime(buf, offset, ie)\n", indent))
				sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent))
				sb.WriteString(fmt.Sprintf("%s\tieLen := ieBefore - offset\n", indent))
				sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependUvarint(buf, offset, uint64(ieLen))\n", indent))
			case innerRinfo.Type == reflect.TypeOf(time.Duration(0)):
				sb.WriteString(fmt.Sprintf("%s\tieBefore := offset\n", indent))
				sb.WriteString(fmt.Sprintf("%s\toffset, err = amino.PrependDuration(buf, offset, ie)\n", indent))
				sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent))
				sb.WriteString(fmt.Sprintf("%s\tieLen := ieBefore - offset\n", indent))
				sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependUvarint(buf, offset, uint64(ieLen))\n", indent))
			case innerRinfo.Type.Kind() == reflect.Struct:
				sb.WriteString(fmt.Sprintf("%s\tieBefore := offset\n", indent))
				sb.WriteString(fmt.Sprintf("%s\toffset, err = ie.MarshalBinary2(cdc, buf, offset)\n", indent))
				sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent))
				sb.WriteString(fmt.Sprintf("%s\tieLen := ieBefore - offset\n", indent))
				sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependUvarint(buf, offset, uint64(ieLen))\n", indent))
			default:
				// String, []byte: writePrimitiveEncode includes own length prefix.
				ctx.writePrimitiveEncode(sb, "ie", innerEinfo, fopts, indent+"\t")
			}
			sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n", indent))
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
		}
	default:
		ctx.writePrimitiveEncode(sb, accessor, einfo, fopts, indent)
	}
}

func (ctx *P3Context2) writeInterfaceFieldMarshal(sb *strings.Builder, accessor string, fnum uint32, indent string) {
	sb.WriteString(fmt.Sprintf("%sif %s != nil {\n", indent, accessor))
	sb.WriteString(fmt.Sprintf("%s\tanyBz, err := cdc.MarshalAny(%s)\n", indent, accessor))
	sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent))
	sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependByteSlice(buf, offset, anyBz)\n", indent))
	sb.WriteString(fmt.Sprintf("%s\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", indent, fnum))
	sb.WriteString(fmt.Sprintf("%s}\n", indent))
}

// === Primitive Encoding ===

func (ctx *P3Context2) writePrimitiveEncode(sb *strings.Builder, accessor string, info *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	rinfo := info.ReprType
	rt := rinfo.Type
	kind := rt.Kind()

	switch kind {
	case reflect.Bool:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependBool(buf, offset, bool(%s))\n", indent, accessor))
	case reflect.Int8, reflect.Int16:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependVarint(buf, offset, int64(%s))\n", indent, accessor))
	case reflect.Int32:
		if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependInt32(buf, offset, int32(%s))\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependVarint(buf, offset, int64(%s))\n", indent, accessor))
		}
	case reflect.Int64:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependInt64(buf, offset, int64(%s))\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependVarint(buf, offset, int64(%s))\n", indent, accessor))
		}
	case reflect.Int:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependInt64(buf, offset, int64(%s))\n", indent, accessor))
		} else if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependInt32(buf, offset, int32(%s))\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependVarint(buf, offset, int64(%s))\n", indent, accessor))
		}
	case reflect.Uint8:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor))
	case reflect.Uint16:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor))
	case reflect.Uint32:
		if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUint32(buf, offset, uint32(%s))\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor))
		}
	case reflect.Uint64:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUint64(buf, offset, uint64(%s))\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor))
		}
	case reflect.Uint:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUint64(buf, offset, uint64(%s))\n", indent, accessor))
		} else if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUint32(buf, offset, uint32(%s))\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor))
		}
	case reflect.Float32:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFloat32(buf, offset, float32(%s))\n", indent, accessor))
	case reflect.Float64:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependFloat64(buf, offset, float64(%s))\n", indent, accessor))
	case reflect.String:
		sb.WriteString(fmt.Sprintf("%soffset = amino.PrependString(buf, offset, string(%s))\n", indent, accessor))
	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependByteSlice(buf, offset, %s)\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive slice element kind %v\n", indent, rt.Elem().Kind()))
		}
	case reflect.Array:
		if rt.Elem().Kind() == reflect.Uint8 {
			sb.WriteString(fmt.Sprintf("%soffset = amino.PrependByteSlice(buf, offset, %s[:])\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive array element kind %v\n", indent, rt.Elem().Kind()))
		}
	default:
		sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive kind %v\n", indent, kind))
	}
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

package genproto2

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// === Entry Point ===

func (ctx *P3Context2) generateSize(sb *strings.Builder, info *amino.TypeInfo) error {
	tname := typeName(info)
	if tname == "" {
		return nil
	}
	if info.Type.Kind() != reflect.Struct && !info.IsAminoMarshaler {
		return nil
	}

	sb.WriteString(fmt.Sprintf("func (goo %s) SizeBinary2(cdc *amino.Codec) int {\n", tname))
	sb.WriteString("\tvar s int\n")

	if info.IsAminoMarshaler {
		sb.WriteString("\trepr, err := goo.MarshalAmino()\n")
		sb.WriteString("\tif err != nil {\n\t\tpanic(err)\n\t}\n")
		rinfo := info.ReprType
		ctx.writeReprSize(sb, rinfo)
		sb.WriteString("\treturn s\n")
		sb.WriteString("}\n\n")
		return nil
	}

	ctx.writeStructSizeBody(sb, info, "goo")

	sb.WriteString("\treturn s\n")
	sb.WriteString("}\n\n")
	return nil
}

// === AminoMarshaler Repr Handling ===

func (ctx *P3Context2) writeReprSize(sb *strings.Builder, rinfo *amino.TypeInfo) {
	rt := rinfo.Type
	fopts := amino.FieldOptions{}

	switch {
	case rt.Kind() == reflect.Struct:
		ctx.writeStructSizeBody(sb, rinfo, "repr")

	case isListType(rt):
		if !rinfo.IsStructOrUnpacked(fopts) {
			// Packed slice repr: wrapped in implicit struct field 1.
			ctx.writePackedSliceReprSize(sb, rinfo)
		} else {
			ctx.writeSliceReprSize(sb, rinfo)
		}

	default:
		// Primitive repr wrapped in implicit struct field 1.
		typ3 := rinfo.GetTyp3(fopts)
		fks := fieldKeySize(1, typ3)
		sb.WriteString("\t{\n")
		sb.WriteString(fmt.Sprintf("\t\tvs := %s\n", ctx.primitiveValueSizeExpr("repr", rinfo, fopts)))
		// Match writeFieldIfNotEmpty: skip if value is just 0x00.
		sb.WriteString(fmt.Sprintf("\t\tif vs > 0 {\n"))
		sb.WriteString(fmt.Sprintf("\t\t\ts += %d + vs\n", fks))
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
	}
}

func (ctx *P3Context2) writePackedSliceReprSize(sb *strings.Builder, info *amino.TypeInfo) {
	einfo := info.Elem
	fopts := amino.FieldOptions{}
	beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8

	sb.WriteString("\tif len(repr) > 0 {\n")
	sb.WriteString("\t\tvar cs int\n")
	if beOptionByte {
		// Each element is 1 byte.
		sb.WriteString("\t\tcs = len(repr)\n")
	} else {
		sb.WriteString("\t\tfor _, elem := range repr {\n")
		if einfo.IsAminoMarshaler {
			sb.WriteString("\t\t\telemRepr, err := elem.MarshalAmino()\n")
			sb.WriteString("\t\t\tif err != nil {\n\t\t\t\tpanic(err)\n\t\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elemRepr", einfo.ReprType, fopts)))
		} else {
			sb.WriteString(fmt.Sprintf("\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elem", einfo, fopts)))
		}
		sb.WriteString("\t\t}\n")
	}
	// Field 1 key + ByteSlice(packed data).
	fks := fieldKeySize(1, amino.Typ3ByteLength)
	sb.WriteString(fmt.Sprintf("\t\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", fks))
	sb.WriteString("\t}\n")
}

func (ctx *P3Context2) writeSliceReprSize(sb *strings.Builder, info *amino.TypeInfo) {
	einfo := info.Elem
	fopts := amino.FieldOptions{}
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength {
		// Packed form: single length-prefixed block.
		sb.WriteString("\tif len(repr) > 0 {\n")
		sb.WriteString("\t\tvar cs int\n")
		sb.WriteString("\t\tfor _, elem := range repr {\n")
		if einfo.IsAminoMarshaler {
			sb.WriteString("\t\t\telemRepr, err := elem.MarshalAmino()\n")
			sb.WriteString("\t\t\tif err != nil {\n\t\t\t\tpanic(err)\n\t\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elemRepr", einfo.ReprType, fopts)))
		} else {
			sb.WriteString(fmt.Sprintf("\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elem", einfo, fopts)))
		}
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\ts += amino.UvarintSize(uint64(cs)) + cs\n")
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field entries with field number 1.
		fks := fieldKeySize(1, amino.Typ3ByteLength)
		sb.WriteString("\tfor _, elem := range repr {\n")
		if einfo.Type.Kind() == reflect.Struct {
			sb.WriteString(fmt.Sprintf("\t\tcs := elem.SizeBinary2(cdc)\n"))
			sb.WriteString(fmt.Sprintf("\t\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", fks))
		} else {
			sb.WriteString(fmt.Sprintf("\t\tvs := %s\n", ctx.primitiveValueSizeExpr("elem", einfo, fopts)))
			sb.WriteString(fmt.Sprintf("\t\ts += %d + vs\n", fks))
		}
		sb.WriteString("\t}\n")
	}
}

// === Struct Fields ===

func (ctx *P3Context2) writeStructSizeBody(sb *strings.Builder, info *amino.TypeInfo, recv string) {
	for _, field := range info.Fields {
		ctx.writeFieldSize(sb, field, recv)
	}
}

func (ctx *P3Context2) writeFieldSize(sb *strings.Builder, field amino.FieldInfo, recv string) {
	finfo := field.TypeInfo
	fname := field.Name
	fnum := field.BinFieldNum
	fopts := field.FieldOptions
	ftype := field.Type
	isPtr := ftype.Kind() == reflect.Ptr

	accessor := fmt.Sprintf("%s.%s", recv, fname)

	if field.UnpackedList {
		ctx.writeUnpackedListSize(sb, accessor, finfo, fopts)
		return
	}

	if isPtr {
		sb.WriteString(fmt.Sprintf("\tif %s != nil {\n", accessor))
		derefAccessor := fmt.Sprintf("(*%s)", accessor)
		if !field.WriteEmpty {
			zeroCheck := ctx.zeroCheck(derefAccessor, finfo, fopts)
			if zeroCheck != "" {
				sb.WriteString(fmt.Sprintf("\t\tif %s {\n", zeroCheck))
				ctx.writeFieldValueSize(sb, derefAccessor, fnum, finfo, fopts, true, "\t\t\t")
				sb.WriteString("\t\t}\n")
			} else {
				ctx.writeFieldValueSize(sb, derefAccessor, fnum, finfo, fopts, true, "\t\t")
			}
		} else {
			ctx.writeFieldValueSize(sb, derefAccessor, fnum, finfo, fopts, true, "\t\t")
		}
		sb.WriteString("\t}\n")
		return
	}

	// Handle AminoMarshaler fields: convert to repr, then size repr.
	if finfo.IsAminoMarshaler && finfo.Type.Kind() != reflect.Struct {
		origZeroCheck := ctx.zeroCheckOriginal(accessor, finfo)
		if origZeroCheck != "" && !field.WriteEmpty {
			sb.WriteString(fmt.Sprintf("\tif %s {\n", origZeroCheck))
			sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn 0\n\t\t}\n")
			ctx.writeFieldValueSize(sb, "repr", fnum, finfo.ReprType, fopts, false, "\t\t")
			sb.WriteString("\t}\n")
		} else {
			sb.WriteString(fmt.Sprintf("\t{\n"))
			sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn 0\n\t\t}\n")
			ctx.writeFieldValueSize(sb, "repr", fnum, finfo.ReprType, fopts, field.WriteEmpty, "\t\t")
			sb.WriteString("\t}\n")
		}
		return
	}
	if finfo.IsAminoMarshaler && finfo.Type.Kind() == reflect.Struct {
		sb.WriteString(fmt.Sprintf("\t{\n"))
		sb.WriteString(fmt.Sprintf("\t\trepr, err := %s.MarshalAmino()\n", accessor))
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn 0\n\t\t}\n")
		rinfo := finfo.ReprType
		if rinfo.Type.Kind() == reflect.Struct {
			sb.WriteString("\t\tcs := repr.SizeBinary2(cdc)\n")
			sb.WriteString(fmt.Sprintf("\t\ts += amino.UvarintSize(uint64(%d<<3|%d))\n", fnum, finfo.GetTyp3(fopts)))
			sb.WriteString("\t\ts += amino.UvarintSize(uint64(cs))\n")
			sb.WriteString("\t\ts += cs\n")
		} else {
			ctx.writeFieldValueSize(sb, "repr", fnum, rinfo, fopts, field.WriteEmpty, "\t\t")
		}
		sb.WriteString("\t}\n")
		return
	}

	if !field.WriteEmpty {
		zeroCheck := ctx.zeroCheck(accessor, finfo, fopts)
		if zeroCheck != "" {
			sb.WriteString(fmt.Sprintf("\tif %s {\n", zeroCheck))
			ctx.writeFieldValueSize(sb, accessor, fnum, finfo, fopts, false, "\t\t")
			sb.WriteString("\t}\n")
			return
		}
	}

	ctx.writeFieldValueSize(sb, accessor, fnum, finfo, fopts, field.WriteEmpty, "\t")
}

func (ctx *P3Context2) writeFieldValueSize(sb *strings.Builder, accessor string, fnum uint32, finfo *amino.TypeInfo, fopts amino.FieldOptions, writeEmpty bool, indent string) {
	typ3 := finfo.GetTyp3(fopts)
	rinfo := finfo.ReprType
	fks := fieldKeySize(fnum, typ3)

	switch {
	case rinfo.Type == reflect.TypeOf(time.Time{}):
		// Time: encode as struct with seconds + nanos fields.
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tcs := amino.TimeSize(%s)\n", indent, accessor))
		ctx.writeByteFieldSizeCheck(sb, fks, writeEmpty, indent)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tcs := amino.DurationSize(%s)\n", indent, accessor))
		ctx.writeByteFieldSizeCheck(sb, fks, writeEmpty, indent)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type.Kind() == reflect.Struct:
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tcs := %s.SizeBinary2(cdc)\n", indent, accessor))
		ctx.writeByteFieldSizeCheck(sb, fks, writeEmpty, indent)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type.Kind() == reflect.Interface:
		sb.WriteString(fmt.Sprintf("%sif %s != nil {\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\tanyBz, err := cdc.MarshalAny(%s)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\tpanic(err)\n%s\t}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s\ts += %d + amino.UvarintSize(uint64(len(anyBz))) + len(anyBz)\n", indent, fks))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	case rinfo.Type.Kind() == reflect.String:
		sb.WriteString(fmt.Sprintf("%ss += %d + amino.UvarintSize(uint64(len(%s))) + len(%s)\n",
			indent, fks, accessor, accessor))

	case rinfo.Type.Kind() == reflect.Slice && rinfo.Type.Elem().Kind() == reflect.Uint8:
		// []byte
		sb.WriteString(fmt.Sprintf("%ss += %d + amino.ByteSliceSize(%s)\n", indent, fks, accessor))

	case rinfo.Type.Kind() == reflect.Array && rinfo.Type.Elem().Kind() == reflect.Uint8:
		// [N]byte
		sb.WriteString(fmt.Sprintf("%ss += %d + amino.UvarintSize(uint64(len(%s))) + len(%s)\n",
			indent, fks, accessor, accessor))

	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Packed list (non-byte elements).
		sb.WriteString(fmt.Sprintf("%s{\n", indent))
		sb.WriteString(fmt.Sprintf("%s\tvar cs int\n", indent))
		einfo := finfo.Elem
		ert := rinfo.Type.Elem()
		eFopts := fopts
		if einfo != nil && einfo.ReprType.Type.Kind() == reflect.Uint8 {
			// List of (repr) bytes: each element is 1 byte.
			sb.WriteString(fmt.Sprintf("%s\tcs = len(%s)\n", indent, accessor))
		} else if einfo != nil {
			sizeExpr := ctx.primitiveValueSizeExpr("e", einfo, eFopts)
			elemUsed := strings.Contains(sizeExpr, "e")
			if ert.Kind() == reflect.Ptr {
				sb.WriteString(fmt.Sprintf("%s\tfor _, e := range %s {\n", indent, accessor))
				sb.WriteString(fmt.Sprintf("%s\t\tif e == nil {\n%s\t\t\tcs += %s\n%s\t\t} else {\n",
					indent, indent, ctx.primitiveValueSizeExpr(fmt.Sprintf("*new(%s)", ctx.goTypeName(ert.Elem())), einfo, eFopts), indent))
				sb.WriteString(fmt.Sprintf("%s\t\t\tcs += %s\n", indent, ctx.primitiveValueSizeExpr("(*e)", einfo, eFopts)))
				sb.WriteString(fmt.Sprintf("%s\t\t}\n", indent))
				sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
			} else if !elemUsed {
				sb.WriteString(fmt.Sprintf("%s\tcs = len(%s) * (%s)\n", indent, accessor, sizeExpr))
			} else {
				sb.WriteString(fmt.Sprintf("%s\tfor _, e := range %s {\n", indent, accessor))
				sb.WriteString(fmt.Sprintf("%s\t\tcs += %s\n", indent, sizeExpr))
				sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
			}
		}
		// For packed lists, always include: the outer len check ensures non-empty.
		ctx.writeByteFieldSizeCheck(sb, fks, true, indent)
		sb.WriteString(fmt.Sprintf("%s}\n", indent))

	default:
		// Primitive types.
		sb.WriteString(fmt.Sprintf("%ss += %d + %s\n",
			indent, fks, ctx.primitiveValueSizeExpr(accessor, rinfo, fopts)))
	}
}

// writeByteFieldSizeCheck writes the size accumulation for a ByteLength field
// after `cs` (content size) has been computed.
func (ctx *P3Context2) writeByteFieldSizeCheck(sb *strings.Builder, fks int, writeEmpty bool, indent string) {
	if !writeEmpty {
		sb.WriteString(fmt.Sprintf("%s\tif cs > 0 {\n", indent))
		sb.WriteString(fmt.Sprintf("%s\t\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", indent, fks))
		sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
	} else {
		sb.WriteString(fmt.Sprintf("%s\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", indent, fks))
	}
}

// === List / Repeated Field Sizing ===

func (ctx *P3Context2) writeUnpackedListSize(sb *strings.Builder, accessor string, finfo *amino.TypeInfo, fopts amino.FieldOptions) {
	ert := finfo.Type.Elem()
	einfo := finfo.Elem
	if einfo == nil {
		return
	}

	// beOptionByte: when element repr is uint8, amino encodes each element
	// as a raw byte rather than a varint (packed as a byte string).
	beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength || beOptionByte {
		// Packed form: field key (fks bytes) + ByteSlice(packed content).
		fks := fieldKeySize(fopts.BinFieldNum, amino.Typ3ByteLength)
		sb.WriteString(fmt.Sprintf("\tif len(%s) > 0 {\n", accessor))
		sb.WriteString("\t\tvar cs int\n")
		if beOptionByte {
			// Byte elements: 1 byte each.
			sb.WriteString(fmt.Sprintf("\t\tcs = len(%s)\n", accessor))
		} else {
			sb.WriteString(fmt.Sprintf("\t\tfor _, elem := range %s {\n", accessor))
			if ert.Kind() == reflect.Ptr {
				sb.WriteString(fmt.Sprintf("\t\t\tif elem == nil {\n"))
				sb.WriteString(fmt.Sprintf("\t\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr(fmt.Sprintf("*new(%s)", ctx.goTypeName(ert.Elem())), einfo, fopts)))
				sb.WriteString(fmt.Sprintf("\t\t\t} else {\n"))
				sb.WriteString(fmt.Sprintf("\t\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("(*elem)", einfo, fopts)))
				sb.WriteString(fmt.Sprintf("\t\t\t}\n"))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elem", einfo, fopts)))
			}
			sb.WriteString("\t\t}\n")
		}
		sb.WriteString(fmt.Sprintf("\t\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", fks))
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field key per element.
		fks := fieldKeySize(fopts.BinFieldNum, amino.Typ3ByteLength)
		ertIsPointer := ert.Kind() == reflect.Ptr
		writeImplicit := isListType(einfo.Type) &&
			einfo.Elem != nil &&
			einfo.Elem.ReprType.Type.Kind() != reflect.Uint8 &&
			einfo.Elem.ReprType.GetTyp3(fopts) != amino.Typ3ByteLength

		sb.WriteString(fmt.Sprintf("\tfor _, elem := range %s {\n", accessor))

		if ertIsPointer {
			ertIsStruct := einfo.ReprType.Type.Kind() == reflect.Struct
			sb.WriteString("\t\tif elem == nil {\n")
			if ertIsStruct && !fopts.NilElements {
				// Match amino: nil struct pointers in lists panic (size is never called if marshal would error).
				sb.WriteString("\t\t\tpanic(\"nil struct pointers in lists not supported unless nil_elements field tag is also set\")\n")
			} else {
				sb.WriteString(fmt.Sprintf("\t\t\ts += %d + 1\n", fks)) // field key + 0x00 byte
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
			// Interface element: size via MarshalAny.
			sb.WriteString(fmt.Sprintf("%sif %s != nil {\n", extraIndent, elemAccessor))
			sb.WriteString(fmt.Sprintf("%s\tanyBz, err := cdc.MarshalAny(%s)\n", extraIndent, elemAccessor))
			sb.WriteString(fmt.Sprintf("%s\tif err != nil {\n%s\t\tpanic(err)\n%s\t}\n", extraIndent, extraIndent, extraIndent))
			sb.WriteString(fmt.Sprintf("%s\ts += %d + amino.UvarintSize(uint64(len(anyBz))) + len(anyBz)\n", extraIndent, fks))
			sb.WriteString(fmt.Sprintf("%s} else {\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\ts += %d + 1\n", extraIndent, fks)) // field key + 0x00
			sb.WriteString(fmt.Sprintf("%s}\n", extraIndent))
		} else if writeImplicit {
			// Nested list: implicit struct wrapping.
			// If inner list is empty, implicit struct is empty (no inner field).
			ifks := fieldKeySize(1, amino.Typ3ByteLength)
			sb.WriteString(fmt.Sprintf("%svar ics int\n", extraIndent))
			ctx.writeListContentSize(sb, elemAccessor, einfo, fopts, extraIndent)
			sb.WriteString(fmt.Sprintf("%svar iss int\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%sif ics > 0 {\n", extraIndent))
			sb.WriteString(fmt.Sprintf("%s\tiss = %d + amino.UvarintSize(uint64(ics)) + ics\n", extraIndent, ifks))
			sb.WriteString(fmt.Sprintf("%s}\n", extraIndent))
			// outer = field key + ByteSlice(iss)
			sb.WriteString(fmt.Sprintf("%ss += %d + amino.UvarintSize(uint64(iss)) + iss\n", extraIndent, fks))
		} else if einfo.ReprType.Type.Kind() == reflect.Struct ||
			einfo.ReprType.Type == reflect.TypeOf(time.Duration(0)) ||
			(isListType(einfo.ReprType.Type) && einfo.ReprType.Type.Elem().Kind() != reflect.Uint8) {
			// Struct/Duration/nested-list element: length-prefixed.
			sb.WriteString(fmt.Sprintf("%scs := %s\n", extraIndent, ctx.elementContentSizeExpr(elemAccessor, einfo, fopts)))
			sb.WriteString(fmt.Sprintf("%ss += %d + amino.UvarintSize(uint64(cs)) + cs\n", extraIndent, fks))
		} else {
			// Non-struct ByteLength element (string, []byte): includes own length prefix.
			sb.WriteString(fmt.Sprintf("%svs := %s\n", extraIndent, ctx.primitiveValueSizeExpr(elemAccessor, einfo, fopts)))
			sb.WriteString(fmt.Sprintf("%ss += %d + vs\n", extraIndent, fks))
		}

		if ertIsPointer {
			sb.WriteString("\t\t}\n")
		}
		sb.WriteString("\t}\n")
	}
}

// writeListContentSize computes packed list content size into variable "ics".
func (ctx *P3Context2) writeListContentSize(sb *strings.Builder, accessor string, info *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	einfo := info.Elem
	ert := info.Type.Elem()
	typ3 := einfo.GetTyp3(fopts)

	if typ3 != amino.Typ3ByteLength {
		sizeExpr := ctx.primitiveValueSizeExpr("e", einfo, fopts)
		elemUsed := strings.Contains(sizeExpr, "e")
		if ert.Kind() == reflect.Ptr {
			sb.WriteString(fmt.Sprintf("%sfor _, e := range %s {\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%s\tif e == nil {\n%s\t\tics += %s\n%s\t} else {\n",
				indent, indent, ctx.primitiveValueSizeExpr(fmt.Sprintf("*new(%s)", ctx.goTypeName(ert.Elem())), einfo, fopts), indent))
			sb.WriteString(fmt.Sprintf("%s\t\tics += %s\n%s\t}\n", indent, ctx.primitiveValueSizeExpr("(*e)", einfo, fopts), indent))
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
		} else if !elemUsed {
			sb.WriteString(fmt.Sprintf("%sics = len(%s) * (%s)\n", indent, accessor, sizeExpr))
		} else {
			sb.WriteString(fmt.Sprintf("%sfor _, e := range %s {\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%s\tics += %s\n", indent, sizeExpr))
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
		}
	} else {
		efks := fieldKeySize(1, amino.Typ3ByteLength)
		sb.WriteString(fmt.Sprintf("%sfor _, e := range %s {\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%s\tecs := %s\n", indent, ctx.elementContentSizeExpr("e", einfo, fopts)))
		sb.WriteString(fmt.Sprintf("%s\tics += %d + amino.UvarintSize(uint64(ecs)) + ecs\n", indent, efks))
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	}
}

// elementContentSizeExpr returns a Go expression for the bare content size of an element
// (before ByteSlice length-prefixing).
func (ctx *P3Context2) elementContentSizeExpr(accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions) string {
	rinfo := einfo.ReprType
	switch {
	case rinfo.Type == reflect.TypeOf(time.Time{}):
		return fmt.Sprintf("amino.TimeSize(%s)", accessor)
	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		return fmt.Sprintf("amino.DurationSize(%s)", accessor)
	case rinfo.Type.Kind() == reflect.Struct:
		return fmt.Sprintf("%s.SizeBinary2(cdc)", accessor)
	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Nested list: use cdc.MarshalSizeOf or marshal+len.
		return fmt.Sprintf("func() int { bz, _ := cdc.Marshal(%s); return len(bz) }()", accessor)
	default:
		return ctx.primitiveValueSizeExpr(accessor, einfo, fopts)
	}
}

// === Primitive Size Expressions ===

// primitiveValueSizeExpr returns a Go expression for the encoded size of a primitive value.
func (ctx *P3Context2) primitiveValueSizeExpr(accessor string, info *amino.TypeInfo, fopts amino.FieldOptions) string {
	rinfo := info.ReprType
	rt := rinfo.Type
	kind := rt.Kind()

	switch kind {
	case reflect.Bool:
		return "1"
	case reflect.Int8, reflect.Int16:
		return fmt.Sprintf("amino.VarintSize(int64(%s))", accessor)
	case reflect.Int32:
		if fopts.BinFixed32 {
			return "4"
		}
		return fmt.Sprintf("amino.VarintSize(int64(%s))", accessor)
	case reflect.Int64:
		if fopts.BinFixed64 {
			return "8"
		}
		return fmt.Sprintf("amino.VarintSize(int64(%s))", accessor)
	case reflect.Int:
		if fopts.BinFixed64 {
			return "8"
		}
		if fopts.BinFixed32 {
			return "4"
		}
		return fmt.Sprintf("amino.VarintSize(int64(%s))", accessor)
	case reflect.Uint8:
		return fmt.Sprintf("amino.UvarintSize(uint64(%s))", accessor)
	case reflect.Uint16:
		return fmt.Sprintf("amino.UvarintSize(uint64(%s))", accessor)
	case reflect.Uint32:
		if fopts.BinFixed32 {
			return "4"
		}
		return fmt.Sprintf("amino.UvarintSize(uint64(%s))", accessor)
	case reflect.Uint64:
		if fopts.BinFixed64 {
			return "8"
		}
		return fmt.Sprintf("amino.UvarintSize(uint64(%s))", accessor)
	case reflect.Uint:
		if fopts.BinFixed64 {
			return "8"
		}
		if fopts.BinFixed32 {
			return "4"
		}
		return fmt.Sprintf("amino.UvarintSize(uint64(%s))", accessor)
	case reflect.Float32:
		return "4"
	case reflect.Float64:
		return "8"
	case reflect.String:
		return fmt.Sprintf("amino.UvarintSize(uint64(len(%s))) + len(%s)", accessor, accessor)
	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			return fmt.Sprintf("amino.ByteSliceSize(%s)", accessor)
		}
		return "0 /* TODO: unsupported */"
	case reflect.Array:
		if rt.Elem().Kind() == reflect.Uint8 {
			return fmt.Sprintf("amino.UvarintSize(uint64(len(%s))) + len(%s)", accessor, accessor)
		}
		return "0 /* TODO: unsupported */"
	default:
		return "0 /* TODO: unsupported */"
	}
}

// fieldKeySize returns the byte size of a protobuf field key.
func fieldKeySize(fnum uint32, typ3 amino.Typ3) int {
	return amino.UvarintSize(uint64(fnum)<<3 | uint64(typ3))
}

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
	fmt.Fprintf(sb, "func (goo %s) SizeBinary2(cdc *amino.Codec) (int, error) {\n", tname)
	sb.WriteString("\tvar s int\n")

	if info.IsAminoMarshaler {
		sb.WriteString("\trepr, err := goo.MarshalAmino()\n")
		sb.WriteString("\tif err != nil {\n\t\treturn 0, err\n\t}\n")
		rinfo := info.ReprType
		ctx.writeReprSize(sb, rinfo)
		sb.WriteString("\treturn s, nil\n")
		sb.WriteString("}\n\n")
		return nil
	}

	// Handle struct types.
	if info.Type.Kind() == reflect.Struct {
		ctx.writeStructSizeBody(sb, info, "goo")
		sb.WriteString("\treturn s, nil\n")
		sb.WriteString("}\n\n")
		return nil
	}

	// Handle non-struct primitive types (e.g. `type StringValue string`).
	// Encoded as implicit struct with a single field number 1.
	sb.WriteString("\trepr := goo\n")
	ctx.writeReprSize(sb, info)
	sb.WriteString("\treturn s, nil\n")
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
		fmt.Fprintf(sb, "\t\tvs := %s\n", ctx.primitiveValueSizeExpr("repr", rinfo, fopts))
		// Match writeFieldIfNotEmpty: skip if value is just 0x00.
		fmt.Fprintf(sb, "\t\tif vs > 0 {\n")
		fmt.Fprintf(sb, "\t\t\ts += %d + vs\n", fks)
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
			sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn 0, err\n\t\t\t}\n")
			fmt.Fprintf(sb, "\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elemRepr", einfo.ReprType, fopts))
		} else {
			fmt.Fprintf(sb, "\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elem", einfo, fopts))
		}
		sb.WriteString("\t\t}\n")
	}
	// Field 1 key + ByteSlice(packed data).
	fks := fieldKeySize(1, amino.Typ3ByteLength)
	fmt.Fprintf(sb, "\t\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", fks)
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
			sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn 0, err\n\t\t\t}\n")
			fmt.Fprintf(sb, "\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elemRepr", einfo.ReprType, fopts))
		} else {
			fmt.Fprintf(sb, "\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elem", einfo, fopts))
		}
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\ts += amino.UvarintSize(uint64(cs)) + cs\n")
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field entries with field number 1.
		fks := fieldKeySize(1, amino.Typ3ByteLength)
		sb.WriteString("\tfor _, elem := range repr {\n")
		if einfo.Type.Kind() == reflect.Struct {
			fmt.Fprintf(sb, "\t\tcs, err := elem.SizeBinary2(cdc)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn 0, err\n\t\t}\n")
			fmt.Fprintf(sb, "\t\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", fks)
		} else {
			fmt.Fprintf(sb, "\t\tvs := %s\n", ctx.primitiveValueSizeExpr("elem", einfo, fopts))
			fmt.Fprintf(sb, "\t\ts += %d + vs\n", fks)
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
		fmt.Fprintf(sb, "\tif %s != nil {\n", accessor)
		derefAccessor := fmt.Sprintf("(*%s)", accessor)
		if !field.WriteEmpty {
			zeroCheck := ctx.zeroCheck(derefAccessor, finfo, fopts)
			if zeroCheck != "" {
				fmt.Fprintf(sb, "\t\tif %s {\n", zeroCheck)
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
			fmt.Fprintf(sb, "\tif %s {\n", origZeroCheck)
			fmt.Fprintf(sb, "\t\trepr, err := %s.MarshalAmino()\n", accessor)
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn 0, err\n\t\t}\n")
			ctx.writeFieldValueSize(sb, "repr", fnum, finfo.ReprType, fopts, false, "\t\t")
			sb.WriteString("\t}\n")
		} else {
			fmt.Fprintf(sb, "\t{\n")
			fmt.Fprintf(sb, "\t\trepr, err := %s.MarshalAmino()\n", accessor)
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn 0, err\n\t\t}\n")
			ctx.writeFieldValueSize(sb, "repr", fnum, finfo.ReprType, fopts, field.WriteEmpty, "\t\t")
			sb.WriteString("\t}\n")
		}
		return
	}
	if finfo.IsAminoMarshaler && finfo.Type.Kind() == reflect.Struct {
		fmt.Fprintf(sb, "\t{\n")
		fmt.Fprintf(sb, "\t\trepr, err := %s.MarshalAmino()\n", accessor)
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn 0, err\n\t\t}\n")
		rinfo := finfo.ReprType
		if rinfo.Type.Kind() == reflect.Struct {
			sb.WriteString("\t\tcs, err := repr.SizeBinary2(cdc)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn 0, err\n\t\t}\n")
			// Match MarshalBinary2's writeLengthPrefixedField: if !WriteEmpty,
			// skip emitting field key + length when inner content is empty.
			// writeByteFieldSizeCheck takes the OUTER block indent and adds
			// a tab internally for the emitted statements.
			fks := fieldKeySize(fnum, finfo.GetTyp3(fopts))
			ctx.writeByteFieldSizeCheck(sb, fks, field.WriteEmpty, "\t")
		} else {
			ctx.writeFieldValueSize(sb, "repr", fnum, rinfo, fopts, field.WriteEmpty, "\t\t")
		}
		sb.WriteString("\t}\n")
		return
	}

	if !field.WriteEmpty {
		zeroCheck := ctx.zeroCheck(accessor, finfo, fopts)
		if zeroCheck != "" {
			fmt.Fprintf(sb, "\tif %s {\n", zeroCheck)
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
		fmt.Fprintf(sb, "%s{\n", indent)
		fmt.Fprintf(sb, "%s\tcs := amino.TimeSize(%s)\n", indent, accessor)
		ctx.writeByteFieldSizeCheck(sb, fks, writeEmpty, indent)
		fmt.Fprintf(sb, "%s}\n", indent)

	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		fmt.Fprintf(sb, "%s{\n", indent)
		fmt.Fprintf(sb, "%s\tcs := amino.DurationSize(%s)\n", indent, accessor)
		ctx.writeByteFieldSizeCheck(sb, fks, writeEmpty, indent)
		fmt.Fprintf(sb, "%s}\n", indent)

	case rinfo.Type.Kind() == reflect.Struct:
		fmt.Fprintf(sb, "%s{\n", indent)
		fmt.Fprintf(sb, "%s\tcs, err := %s.SizeBinary2(cdc)\n", indent, accessor)
		fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn 0, err\n%s\t}\n", indent, indent, indent)
		ctx.writeByteFieldSizeCheck(sb, fks, writeEmpty, indent)
		fmt.Fprintf(sb, "%s}\n", indent)

	case rinfo.Type.Kind() == reflect.Interface:
		fmt.Fprintf(sb, "%sif %s != nil {\n", indent, accessor)
		fmt.Fprintf(sb, "%s\tcs, err := cdc.SizeAnyBinary2(%s)\n", indent, accessor)
		fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn 0, err\n%s\t}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", indent, fks)
		fmt.Fprintf(sb, "%s}\n", indent)

	case rinfo.Type.Kind() == reflect.String:
		fmt.Fprintf(sb, "%ss += %d + amino.UvarintSize(uint64(len(%s))) + len(%s)\n",
			indent, fks, accessor, accessor)

	case rinfo.Type.Kind() == reflect.Slice && rinfo.Type.Elem().Kind() == reflect.Uint8:
		// []byte
		fmt.Fprintf(sb, "%ss += %d + amino.ByteSliceSize(%s)\n", indent, fks, accessor)

	case rinfo.Type.Kind() == reflect.Array && rinfo.Type.Elem().Kind() == reflect.Uint8:
		// [N]byte
		fmt.Fprintf(sb, "%ss += %d + amino.UvarintSize(uint64(len(%s))) + len(%s)\n",
			indent, fks, accessor, accessor)

	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Packed list (non-byte elements).
		fmt.Fprintf(sb, "%s{\n", indent)
		fmt.Fprintf(sb, "%s\tvar cs int\n", indent)
		einfo := finfo.Elem
		ert := rinfo.Type.Elem()
		eFopts := fopts
		if einfo != nil && einfo.ReprType.Type.Kind() == reflect.Uint8 {
			fmt.Fprintf( // List of (repr) bytes: each element is 1 byte.
				sb, "%s\tcs = len(%s)\n", indent, accessor)
		} else if einfo != nil {
			sizeExpr := ctx.primitiveValueSizeExpr("e", einfo, eFopts)
			elemUsed := strings.Contains(sizeExpr, "e")
			if ert.Kind() == reflect.Ptr {
				fmt.Fprintf(sb, "%s\tfor _, e := range %s {\n", indent, accessor)
				fmt.Fprintf(sb, "%s\t\tif e == nil {\n%s\t\t\tcs += %s\n%s\t\t} else {\n",
					indent, indent, ctx.primitiveValueSizeExpr(fmt.Sprintf("*new(%s)", ctx.goTypeName(ert.Elem())), einfo, eFopts), indent)
				fmt.Fprintf(sb, "%s\t\t\tcs += %s\n", indent, ctx.primitiveValueSizeExpr("(*e)", einfo, eFopts))
				fmt.Fprintf(sb, "%s\t\t}\n", indent)
				fmt.Fprintf(sb, "%s\t}\n", indent)
			} else if !elemUsed {
				fmt.Fprintf(sb, "%s\tcs = len(%s) * (%s)\n", indent, accessor, sizeExpr)
			} else {
				fmt.Fprintf(sb, "%s\tfor _, e := range %s {\n", indent, accessor)
				fmt.Fprintf(sb, "%s\t\tcs += %s\n", indent, sizeExpr)
				fmt.Fprintf(sb, "%s\t}\n", indent)
			}
		}
		// For packed lists, always include: the outer len check ensures non-empty.
		ctx.writeByteFieldSizeCheck(sb, fks, true, indent)
		fmt.Fprintf(sb, "%s}\n", indent)

	default:
		// Primitive types.
		fmt.Fprintf(sb, "%ss += %d + %s\n",
			indent, fks, ctx.primitiveValueSizeExpr(accessor, rinfo, fopts))
	}
}

// writeByteFieldSizeCheck writes the size accumulation for a ByteLength field
// after `cs` (content size) has been computed.
func (ctx *P3Context2) writeByteFieldSizeCheck(sb *strings.Builder, fks int, writeEmpty bool, indent string) {
	if !writeEmpty {
		fmt.Fprintf(sb, "%s\tif cs > 0 {\n", indent)
		fmt.Fprintf(sb, "%s\t\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", indent, fks)
		fmt.Fprintf(sb, "%s\t}\n", indent)
	} else {
		fmt.Fprintf(sb, "%s\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", indent, fks)
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
		fmt.Fprintf(sb, "\tif len(%s) > 0 {\n", accessor)
		sb.WriteString("\t\tvar cs int\n")
		if beOptionByte {
			fmt.Fprintf( // Byte elements: 1 byte each.
				sb, "\t\tcs = len(%s)\n", accessor)
		} else {
			fmt.Fprintf(sb, "\t\tfor _, elem := range %s {\n", accessor)
			if ert.Kind() == reflect.Ptr {
				fmt.Fprintf(sb, "\t\t\tif elem == nil {\n")
				fmt.Fprintf(sb, "\t\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr(fmt.Sprintf("*new(%s)", ctx.goTypeName(ert.Elem())), einfo, fopts))
				fmt.Fprintf(sb, "\t\t\t} else {\n")
				fmt.Fprintf(sb, "\t\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("(*elem)", einfo, fopts))
				fmt.Fprintf(sb, "\t\t\t}\n")
			} else {
				fmt.Fprintf(sb, "\t\t\tcs += %s\n", ctx.primitiveValueSizeExpr("elem", einfo, fopts))
			}
			sb.WriteString("\t\t}\n")
		}
		fmt.Fprintf(sb, "\t\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", fks)
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field key per element.
		fks := fieldKeySize(fopts.BinFieldNum, amino.Typ3ByteLength)
		ertIsPointer := ert.Kind() == reflect.Ptr
		writeImplicit := isListType(einfo.Type) &&
			einfo.Elem != nil &&
			einfo.Elem.ReprType.Type.Kind() != reflect.Uint8 &&
			einfo.Elem.ReprType.GetTyp3(fopts) != amino.Typ3ByteLength
		fmt.Fprintf(sb, "\tfor _, elem := range %s {\n", accessor)

		if ertIsPointer {
			ertIsStruct := einfo.ReprType.Type.Kind() == reflect.Struct
			sb.WriteString("\t\tif elem == nil {\n")
			if ertIsStruct && !fopts.NilElements {
				// Match MarshalBinary2's error surface: Size is a public method
				// and may be called independently, so error rather than panic.
				sb.WriteString("\t\t\treturn 0, errors.New(\"nil struct pointers in lists not supported unless nil_elements field tag is also set\")\n")
			} else {
				fmt.Fprintf(sb, "\t\t\ts += %d + 1\n", fks) // field key + 0x00 byte
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
			fmt.Fprintf( // Interface element: size via SizeAnyBinary2.
				sb, "%sif %s != nil {\n", extraIndent, elemAccessor)
			fmt.Fprintf(sb, "%s\tcs, err := cdc.SizeAnyBinary2(%s)\n", extraIndent, elemAccessor)
			fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn 0, err\n%s\t}\n", extraIndent, extraIndent, extraIndent)
			fmt.Fprintf(sb, "%s\ts += %d + amino.UvarintSize(uint64(cs)) + cs\n", extraIndent, fks)
			fmt.Fprintf(sb, "%s} else {\n", extraIndent)
			fmt.Fprintf(sb, "%s\ts += %d + 1\n", extraIndent, fks)
			fmt. // field key + 0x00
				Fprintf(sb, "%s}\n", extraIndent)
		} else if writeImplicit {
			// Nested list: implicit struct wrapping.
			// If inner list is empty, implicit struct is empty (no inner field).
			ifks := fieldKeySize(1, amino.Typ3ByteLength)
			fmt.Fprintf(sb, "%svar ics int\n", extraIndent)
			ctx.writeListContentSize(sb, elemAccessor, einfo, fopts, extraIndent)
			fmt.Fprintf(sb, "%svar iss int\n", extraIndent)
			fmt.Fprintf(sb, "%sif ics > 0 {\n", extraIndent)
			fmt.Fprintf(sb, "%s\tiss = %d + amino.UvarintSize(uint64(ics)) + ics\n", extraIndent, ifks)
			fmt.Fprintf(sb, "%s}\n", extraIndent)
			fmt.Fprintf( // outer = field key + ByteSlice(iss)
				sb, "%ss += %d + amino.UvarintSize(uint64(iss)) + iss\n", extraIndent, fks)
		} else if einfo.ReprType.Type.Kind() == reflect.Struct ||
			einfo.ReprType.Type == reflect.TypeOf(time.Duration(0)) ||
			(isListType(einfo.ReprType.Type) && einfo.ReprType.Type.Elem().Kind() != reflect.Uint8) {
			// Struct/Duration/nested-list element: length-prefixed.
			ctx.writeElementContentSize(sb, "cs", elemAccessor, einfo, fopts, extraIndent)
			fmt.Fprintf(sb, "%ss += %d + amino.UvarintSize(uint64(cs)) + cs\n", extraIndent, fks)
		} else if einfo.IsAminoMarshaler {
			fmt.Fprintf( // AminoMarshaler element: size the repr value after MarshalAmino.
				sb, "%ser, err := %s.MarshalAmino()\n", extraIndent, elemAccessor)
			fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn 0, err\n%s}\n", extraIndent, extraIndent, extraIndent)
			fmt.Fprintf(sb, "%svs := %s\n", extraIndent, ctx.primitiveValueSizeExpr("er", einfo.ReprType, fopts))
			fmt.Fprintf(sb, "%ss += %d + vs\n", extraIndent, fks)
		} else {
			fmt.Fprintf( // Non-struct ByteLength element (string, []byte): includes own length prefix.
				sb, "%svs := %s\n", extraIndent, ctx.primitiveValueSizeExpr(elemAccessor, einfo, fopts))
			fmt.Fprintf(sb, "%ss += %d + vs\n", extraIndent, fks)
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
			fmt.Fprintf(sb, "%sfor _, e := range %s {\n", indent, accessor)
			fmt.Fprintf(sb, "%s\tif e == nil {\n%s\t\tics += %s\n%s\t} else {\n",
				indent, indent, ctx.primitiveValueSizeExpr(fmt.Sprintf("*new(%s)", ctx.goTypeName(ert.Elem())), einfo, fopts), indent)
			fmt.Fprintf(sb, "%s\t\tics += %s\n%s\t}\n", indent, ctx.primitiveValueSizeExpr("(*e)", einfo, fopts), indent)
			fmt.Fprintf(sb, "%s}\n", indent)
		} else if !elemUsed {
			fmt.Fprintf(sb, "%sics = len(%s) * (%s)\n", indent, accessor, sizeExpr)
		} else {
			fmt.Fprintf(sb, "%sfor _, e := range %s {\n", indent, accessor)
			fmt.Fprintf(sb, "%s\tics += %s\n", indent, sizeExpr)
			fmt.Fprintf(sb, "%s}\n", indent)
		}
	} else {
		efks := fieldKeySize(1, amino.Typ3ByteLength)
		fmt.Fprintf(sb, "%sfor _, e := range %s {\n", indent, accessor)
		ctx.writeElementContentSize(sb, "ecs", "e", einfo, fopts, indent+"\t")
		fmt.Fprintf(sb, "%s\tics += %d + amino.UvarintSize(uint64(ecs)) + ecs\n", indent, efks)
		fmt.Fprintf(sb, "%s}\n", indent)
	}
}

// writeElementContentSize emits Go code that assigns the bare content size
// of an element (before ByteSlice length-prefixing) to varname, propagating
// any SizeBinary2/SizeAnyBinary2 errors via `return 0, err`.
func (ctx *P3Context2) writeElementContentSize(sb *strings.Builder, varname, accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	rinfo := einfo.ReprType
	switch {
	case rinfo.Type == reflect.TypeOf(time.Time{}):
		fmt.Fprintf(sb, "%s%s := amino.TimeSize(%s)\n", indent, varname, accessor)
	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		fmt.Fprintf(sb, "%s%s := amino.DurationSize(%s)\n", indent, varname, accessor)
	case rinfo.Type.Kind() == reflect.Struct:
		fmt.Fprintf(sb, "%s%s, err := %s.SizeBinary2(cdc)\n", indent, varname, accessor)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn 0, err\n%s}\n", indent, indent, indent)
	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Nested list: compute implicit struct content size inline.
		innerEinfo := einfo.Elem
		if innerEinfo == nil {
			fmt.Fprintf(sb, "%sbz, _ := cdc.Marshal(%s)\n", indent, accessor)
			fmt.Fprintf(sb, "%s%s := len(bz)\n", indent, varname)
			return
		}
		innerTyp3 := innerEinfo.GetTyp3(fopts)
		innerRinfo := innerEinfo.ReprType
		ifks := fieldKeySize(1, amino.Typ3ByteLength)
		if innerTyp3 != amino.Typ3ByteLength {
			innerSizeExpr := ctx.primitiveValueSizeExpr("ie", innerEinfo, fopts)
			fmt.Fprintf(sb, "%svar %s int\n", indent, varname)
			fmt.Fprintf(sb, "%sfor _, ie := range %s {\n", indent, accessor)
			fmt.Fprintf(sb, "%s\t%s += %s\n", indent, varname, innerSizeExpr)
			fmt.Fprintf(sb, "%s}\n", indent)
			fmt.Fprintf(sb, "%sif %s > 0 {\n", indent, varname)
			fmt.Fprintf(sb, "%s\t%s = %d + amino.UvarintSize(uint64(%s)) + %s\n", indent, varname, ifks, varname, varname)
			fmt.Fprintf(sb, "%s}\n", indent)
			return
		}
		// ByteLength: repeated field 1 entries.
		fmt.Fprintf(sb, "%svar %s int\n", indent, varname)
		fmt.Fprintf(sb, "%sfor _, ie := range %s {\n", indent, accessor)
		switch {
		case innerRinfo.Type == reflect.TypeOf(time.Time{}):
			fmt.Fprintf(sb, "%s\tts := amino.TimeSize(ie)\n", indent)
			fmt.Fprintf(sb, "%s\t%s += %d + amino.UvarintSize(uint64(ts)) + ts\n", indent, varname, ifks)
		case innerRinfo.Type == reflect.TypeOf(time.Duration(0)):
			fmt.Fprintf(sb, "%s\tds := amino.DurationSize(ie)\n", indent)
			fmt.Fprintf(sb, "%s\t%s += %d + amino.UvarintSize(uint64(ds)) + ds\n", indent, varname, ifks)
		case innerRinfo.Type.Kind() == reflect.Struct:
			fmt.Fprintf(sb, "%s\tes, err := ie.SizeBinary2(cdc)\n", indent)
			fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn 0, err\n%s\t}\n", indent, indent, indent)
			fmt.Fprintf(sb, "%s\t%s += %d + amino.UvarintSize(uint64(es)) + es\n", indent, varname, ifks)
		default:
			fmt.Fprintf(sb, "%s\t%s += %d + %s\n", indent, varname, ifks, ctx.primitiveValueSizeExpr("ie", innerEinfo, fopts))
		}
		fmt.Fprintf(sb, "%s}\n", indent)
	default:
		fmt.Fprintf(sb, "%s%s := %s\n", indent, varname, ctx.primitiveValueSizeExpr(accessor, einfo, fopts))
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
		panic(fmt.Sprintf("genproto2: primitiveValueSizeExpr: unsupported slice element kind %v (type=%v, accessor=%s)", rt.Elem().Kind(), rt, accessor))
	case reflect.Array:
		if rt.Elem().Kind() == reflect.Uint8 {
			return fmt.Sprintf("amino.UvarintSize(uint64(len(%s))) + len(%s)", accessor, accessor)
		}
		panic(fmt.Sprintf("genproto2: primitiveValueSizeExpr: unsupported array element kind %v (type=%v, accessor=%s)", rt.Elem().Kind(), rt, accessor))
	default:
		panic(fmt.Sprintf("genproto2: primitiveValueSizeExpr: unsupported kind %v (type=%v, accessor=%s)", rt.Kind(), rt, accessor))
	}
}

// fieldKeySize returns the byte size of a protobuf field key.
func fieldKeySize(fnum uint32, typ3 amino.Typ3) int {
	return amino.UvarintSize(uint64(fnum)<<3 | uint64(typ3))
}

package genproto2

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// === Entry Point ===

func (ctx *P3Context2) generateUnmarshal(sb *strings.Builder, info *amino.TypeInfo) error {
	tname := typeName(info)
	if tname == "" {
		return nil
	}

	// Only generate for struct types and AminoMarshalers.
	if info.Type.Kind() != reflect.Struct && !info.IsAminoMarshaler {
		return nil
	}

	sb.WriteString(fmt.Sprintf("func (goo *%s) UnmarshalBinary2(cdc *amino.Codec, bz []byte) error {\n", tname))

	if info.IsAminoMarshaler {
		rinfo := info.ReprType
		if err := ctx.writeReprUnmarshal(sb, rinfo); err != nil {
			return err
		}
		sb.WriteString("\treturn goo.UnmarshalAmino(repr)\n")
		sb.WriteString("}\n\n")
		return nil
	}

	ctx.writeStructUnmarshalBody(sb, info, "goo")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")
	return nil
}

// === AminoMarshaler Repr Handling ===

func (ctx *P3Context2) writeReprUnmarshal(sb *strings.Builder, rinfo *amino.TypeInfo) error {
	rt := rinfo.Type
	fopts := amino.FieldOptions{}

	// Struct-or-unpacked repr: decode directly (no implicit struct wrapper).
	if rinfo.IsStructOrUnpacked(fopts) {
		switch rt.Kind() {
		case reflect.Struct:
			sb.WriteString(fmt.Sprintf("\tvar repr %s\n", rt.Name()))
			if rinfo.Registered {
				sb.WriteString("\tif err := repr.UnmarshalBinary2(cdc, bz); err != nil {\n\t\treturn err\n\t}\n")
			} else {
				sb.WriteString("\tif err := cdc.Unmarshal(bz, &repr); err != nil {\n\t\treturn err\n\t}\n")
			}
		case reflect.Slice, reflect.Array:
			ctx.writeSliceReprUnmarshal(sb, rinfo)
		default:
			return fmt.Errorf("unsupported IsStructOrUnpacked repr kind %v", rt.Kind())
		}
		return nil
	}

	// Non-struct-or-unpacked repr: wrapped in implicit struct with field 1.
	if isListType(rt) {
		// Packed slice repr: read field 1 key, then DecodeByteSlice, then decode packed elements.
		einfo := rinfo.Elem
		sb.WriteString(fmt.Sprintf("\tvar repr %s\n", ctx.goTypeName(rt)))
		sb.WriteString("\tif len(bz) > 0 {\n")
		sb.WriteString("\t\t_, _, n, err := amino.DecodeFieldNumberAndTyp3(bz)\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		sb.WriteString("\t\tbz = bz[n:]\n")
		sb.WriteString("\t\tfbz, _, err := amino.DecodeByteSlice(bz)\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		ert := rt.Elem()
		beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8
		if beOptionByte {
			// Each element is a raw byte.
			sb.WriteString("\t\tfor _, b := range fbz {\n")
			if einfo.IsAminoMarshaler {
				sb.WriteString(fmt.Sprintf("\t\t\tvar elem %s\n", ctx.goTypeName(ert)))
				sb.WriteString("\t\t\tif err := elem.UnmarshalAmino(b); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
			} else {
				sb.WriteString(fmt.Sprintf("\t\t\telem := %s(b)\n", ctx.goTypeName(ert)))
			}
			sb.WriteString("\t\t\trepr = append(repr, elem)\n")
			sb.WriteString("\t\t}\n")
		} else {
			sb.WriteString("\t\tfor len(fbz) > 0 {\n")
			if einfo.IsAminoMarshaler {
				reprType := einfo.ReprType.Type
				sb.WriteString(fmt.Sprintf("\t\t\tvar rv %s\n", ctx.goTypeName(reprType)))
				ctx.writePrimitiveDecodeFrom(sb, "rv", einfo.ReprType, fopts, "\t\t\t", "fbz")
				sb.WriteString(fmt.Sprintf("\t\t\tvar elem %s\n", ctx.goTypeName(ert)))
				sb.WriteString("\t\t\tif err := elem.UnmarshalAmino(rv); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
			} else {
				sb.WriteString(fmt.Sprintf("\t\t\tvar elem %s\n", ctx.goTypeName(ert)))
				ctx.writePrimitiveDecodeFrom(sb, "elem", einfo, fopts, "\t\t\t", "fbz")
			}
			sb.WriteString("\t\t\trepr = append(repr, elem)\n")
			sb.WriteString("\t\t}\n")
		}
		sb.WriteString("\t}\n")
	} else {
		// Primitive repr: read field 1 key, then decode value.
		sb.WriteString(fmt.Sprintf("\tvar repr %s\n", ctx.goTypeName(rt)))
		sb.WriteString("\tif len(bz) > 0 {\n")
		sb.WriteString("\t\t_, _, n, err := amino.DecodeFieldNumberAndTyp3(bz)\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		sb.WriteString("\t\tbz = bz[n:]\n")

		switch rt.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sb.WriteString("\t\tv, _, err := amino.DecodeVarint(bz)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\trepr = %s(v)\n", ctx.goTypeName(rt)))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			sb.WriteString("\t\tv, _, err := amino.DecodeUvarint(bz)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\trepr = %s(v)\n", ctx.goTypeName(rt)))
		case reflect.String:
			sb.WriteString("\t\tv, _, err := amino.DecodeString(bz)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\trepr = %s(v)\n", ctx.goTypeName(rt)))
		case reflect.Bool:
			sb.WriteString("\t\tv, _, err := amino.DecodeBool(bz)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\trepr = %s(v)\n", ctx.goTypeName(rt)))
		default:
			return fmt.Errorf("unsupported non-struct repr kind %v", rt.Kind())
		}
		sb.WriteString("\t}\n")
	}
	return nil
}

func (ctx *P3Context2) writeSliceReprUnmarshal(sb *strings.Builder, info *amino.TypeInfo) {
	ert := info.Type.Elem()
	einfo := info.Elem
	fopts := amino.FieldOptions{}
	typ3 := einfo.GetTyp3(fopts)

	sb.WriteString(fmt.Sprintf("\tvar repr %s\n", ctx.goTypeName(info.Type)))

	if typ3 != amino.Typ3ByteLength {
		// Packed: whole bz is the packed data.
		sb.WriteString("\tfor len(bz) > 0 {\n")
		if einfo.IsAminoMarshaler {
			// Decode repr type, then UnmarshalAmino.
			reprType := einfo.ReprType.Type
			sb.WriteString(fmt.Sprintf("\t\tvar rv %s\n", ctx.goTypeName(reprType)))
			ctx.writePrimitiveDecode(sb, "rv", einfo.ReprType, fopts, "\t\t")
			sb.WriteString(fmt.Sprintf("\t\tvar elem %s\n", ctx.goTypeName(ert)))
			sb.WriteString("\t\tif err := elem.UnmarshalAmino(rv); err != nil {\n\t\t\treturn err\n\t\t}\n")
		} else {
			sb.WriteString(fmt.Sprintf("\t\tvar elem %s\n", ctx.goTypeName(ert)))
			ctx.writePrimitiveDecode(sb, "elem", einfo, fopts, "\t\t")
		}
		sb.WriteString("\t\trepr = append(repr, elem)\n")
		sb.WriteString("\t}\n")
	} else {
		// Unpacked: repeated field entries with field number 1.
		sb.WriteString("\tfor len(bz) > 0 {\n")
		sb.WriteString("\t\tfnum, typ3, n, err := amino.DecodeFieldNumberAndTyp3(bz)\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		sb.WriteString("\t\tbz = bz[n:]\n")
		sb.WriteString("\t\tif fnum != 1 {\n")
		sb.WriteString("\t\t\tn, err = amino.SkipField(bz, typ3)\n")
		sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
		sb.WriteString("\t\t\tbz = bz[n:]\n")
		sb.WriteString("\t\t\tcontinue\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString(fmt.Sprintf("\t\tvar elem %s\n", ctx.goTypeName(ert)))
		ctx.writeByteSliceElementDecode(sb, "elem", einfo, fopts, "\t\t")
		sb.WriteString("\t\trepr = append(repr, elem)\n")
		sb.WriteString("\t}\n")
	}
}

// === Struct Fields ===

func (ctx *P3Context2) writeStructUnmarshalBody(sb *strings.Builder, info *amino.TypeInfo, recv string) {
	// Declare index counters for array unpacked-list fields.
	for _, field := range info.Fields {
		if field.UnpackedList && field.Type.Kind() == reflect.Array {
			sb.WriteString(fmt.Sprintf("\tvar %s_idx int\n", field.Name))
		}
	}

	// Initialize non-pointer time.Time fields to amino's "empty" time (1970, not 0001).
	for _, field := range info.Fields {
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			continue
		}
		rinfo := field.TypeInfo.ReprType
		if rinfo.Type == reflect.TypeOf(time.Time{}) {
			sb.WriteString(fmt.Sprintf("\t%s.%s = time.Unix(0, 0).UTC()\n", recv, field.Name))
		}
	}

	sb.WriteString("\tvar lastFieldNum uint32\n")
	sb.WriteString("\tfor len(bz) > 0 {\n")
	sb.WriteString("\t\tfnum, typ3, n, err := amino.DecodeFieldNumberAndTyp3(bz)\n")
	sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
	sb.WriteString("\t\tif fnum < lastFieldNum {\n")
	sb.WriteString("\t\t\treturn fmt.Errorf(\"encountered fieldNum: %v, but we have already seen fnum: %v\", fnum, lastFieldNum)\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t\tlastFieldNum = fnum\n")
	sb.WriteString("\t\tbz = bz[n:]\n")
	sb.WriteString("\t\tswitch fnum {\n")

	for _, field := range info.Fields {
		finfo := field.TypeInfo
		fname := field.Name
		fnum := field.BinFieldNum
		fopts := field.FieldOptions
		ftype := field.Type
		isPtr := ftype.Kind() == reflect.Ptr

		accessor := fmt.Sprintf("%s.%s", recv, fname)

		sb.WriteString(fmt.Sprintf("\t\tcase %d:\n", fnum))

		if field.UnpackedList {
			isArray := ftype.Kind() == reflect.Array
			ctx.writeUnpackedListUnmarshal(sb, accessor, finfo, fopts, "\t\t\t", isArray, fname)
			// Continue consuming repeated entries with the same field number.
			sb.WriteString(fmt.Sprintf("\t\t\tfor len(bz) > 0 {\n"))
			sb.WriteString(fmt.Sprintf("\t\t\t\tvar nextFnum uint32\n"))
			sb.WriteString(fmt.Sprintf("\t\t\t\tnextFnum, _, n, err = amino.DecodeFieldNumberAndTyp3(bz)\n"))
			sb.WriteString(fmt.Sprintf("\t\t\t\tif err != nil {\n\t\t\t\t\treturn err\n\t\t\t\t}\n"))
			sb.WriteString(fmt.Sprintf("\t\t\t\tif nextFnum != %d {\n\t\t\t\t\tbreak\n\t\t\t\t}\n", fnum))
			sb.WriteString(fmt.Sprintf("\t\t\t\tbz = bz[n:]\n"))
			ctx.writeUnpackedListUnmarshal(sb, accessor, finfo, fopts, "\t\t\t\t", isArray, fname)
			sb.WriteString(fmt.Sprintf("\t\t\t}\n"))
		} else if isPtr {
			ctx.writePointerFieldUnmarshal(sb, accessor, ftype, finfo, fopts, "\t\t\t")
		} else if finfo.IsAminoMarshaler {
			// AminoMarshaler field: decode repr, then UnmarshalAmino.
			reprTypeName := ctx.goTypeName(finfo.ReprType.Type)
			sb.WriteString(fmt.Sprintf("\t\t\tvar repr %s\n", reprTypeName))
			ctx.writeFieldUnmarshal(sb, "repr", finfo.ReprType, fopts, "\t\t\t")
			sb.WriteString(fmt.Sprintf("\t\t\tif err := %s.UnmarshalAmino(repr); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n", accessor))
		} else {
			ctx.writeFieldUnmarshal(sb, accessor, finfo, fopts, "\t\t\t")
		}
	}

	sb.WriteString("\t\tdefault:\n")
	sb.WriteString("\t\t\tn, err = amino.SkipField(bz, typ3)\n")
	sb.WriteString("\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
	sb.WriteString("\t\t\tbz = bz[n:]\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n")

	// Initialize pointer fields that were not decoded from wire data.
	// Amino's defaultValue() allocates non-nil pointers for missing fields:
	//   *time.Time   → &time.Unix(0,0).UTC()
	//   *<struct>    → nil (stays nil)
	//   *<non-struct> → new(T) (non-nil pointer to zero value)
	for _, field := range info.Fields {
		ft := field.Type
		if ft.Kind() != reflect.Ptr {
			continue
		}
		ert := ft.Elem()
		accessor := fmt.Sprintf("%s.%s", recv, field.Name)
		if ert == reflect.TypeOf(time.Time{}) {
			sb.WriteString(fmt.Sprintf("\tif %s == nil {\n\t\tv := time.Unix(0, 0).UTC()\n\t\t%s = &v\n\t}\n", accessor, accessor))
		} else if ert.Kind() != reflect.Struct {
			sb.WriteString(fmt.Sprintf("\tif %s == nil {\n\t\t%s = new(%s)\n\t}\n", accessor, accessor, ctx.goTypeName(ert)))
		}
	}
}

func (ctx *P3Context2) writeFieldUnmarshal(sb *strings.Builder, accessor string, finfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	rinfo := finfo.ReprType
	rt := rinfo.Type

	switch {
	case rt == reflect.TypeOf(time.Time{}):
		sb.WriteString(fmt.Sprintf("%s// time.Time (ByteLength)\n", indent))
		sb.WriteString(fmt.Sprintf("%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
		sb.WriteString(fmt.Sprintf("%s%s, _, err = amino.DecodeTime(fbz)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))

	case rt == reflect.TypeOf(time.Duration(0)):
		sb.WriteString(fmt.Sprintf("%s// time.Duration (ByteLength)\n", indent))
		sb.WriteString(fmt.Sprintf("%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
		sb.WriteString(fmt.Sprintf("%s%s, _, err = amino.DecodeDuration(fbz)\n", indent, accessor))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))

	case rt.Kind() == reflect.Struct:
		sb.WriteString(fmt.Sprintf("%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err := %s.UnmarshalBinary2(cdc, fbz); err != nil {\n%s\treturn err\n%s}\n",
			indent, accessor, indent, indent))

	case rt.Kind() == reflect.Interface:
		ctx.writeInterfaceFieldUnmarshal(sb, accessor, indent)

	case rt.Kind() == reflect.String:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeString(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case rt.Kind() == reflect.Slice && rt.Elem().Kind() == reflect.Uint8:
		// []byte
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeByteSlice(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
		sb.WriteString(fmt.Sprintf("%sif len(v) == 0 {\n%s\t%s = nil\n%s} else {\n%s\t%s = v\n%s}\n",
			indent, indent, accessor, indent, indent, accessor, indent))

	case rt.Kind() == reflect.Array && rt.Elem().Kind() == reflect.Uint8:
		// [N]byte
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeByteSlice(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
		sb.WriteString(fmt.Sprintf("%scopy(%s[:], v)\n", indent, accessor))

	case isListType(rt) && rt.Elem().Kind() != reflect.Uint8:
		// Packed list (non-byte elements): decode length-prefixed block, then elements.
		sb.WriteString(fmt.Sprintf("%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
		einfo := finfo.Elem
		ert := rt.Elem()
		if einfo != nil {
			isArray := rt.Kind() == reflect.Array
			arrayLen := 0
			if isArray {
				arrayLen = rt.Len()
				sb.WriteString(fmt.Sprintf("%svar idx int\n", indent))
			}
			if einfo.ReprType.Type.Kind() == reflect.Uint8 {
				// List of (repr) bytes: elements are raw bytes (beOptionByte encoding).
				sb.WriteString(fmt.Sprintf("%sfor _, b := range fbz {\n", indent))
				if isArray {
					sb.WriteString(fmt.Sprintf("%s\tif idx >= %d {\n%s\t\tbreak\n%s\t}\n", indent, arrayLen, indent, indent))
				}
				if einfo.IsAminoMarshaler {
					sb.WriteString(fmt.Sprintf("%s\tvar elem %s\n", indent, ctx.goTypeName(ert)))
					sb.WriteString(fmt.Sprintf("%s\tif err := elem.UnmarshalAmino(b); err != nil {\n%s\t\treturn err\n%s\t}\n", indent, indent, indent))
				} else if ert.Kind() == reflect.Ptr {
					sb.WriteString(fmt.Sprintf("%s\tev := %s(b)\n", indent, ctx.goTypeName(ert.Elem())))
				} else {
					sb.WriteString(fmt.Sprintf("%s\tev := %s(b)\n", indent, ctx.goTypeName(ert)))
				}
				if einfo.IsAminoMarshaler {
					if isArray {
						sb.WriteString(fmt.Sprintf("%s\t%s[idx] = elem\n%s\tidx++\n", indent, accessor, indent))
					} else {
						sb.WriteString(fmt.Sprintf("%s\t%s = append(%s, elem)\n", indent, accessor, accessor))
					}
				} else if ert.Kind() == reflect.Ptr {
					if isArray {
						sb.WriteString(fmt.Sprintf("%s\t%s[idx] = &ev\n%s\tidx++\n", indent, accessor, indent))
					} else {
						sb.WriteString(fmt.Sprintf("%s\t%s = append(%s, &ev)\n", indent, accessor, accessor))
					}
				} else {
					if isArray {
						sb.WriteString(fmt.Sprintf("%s\t%s[idx] = ev\n%s\tidx++\n", indent, accessor, indent))
					} else {
						sb.WriteString(fmt.Sprintf("%s\t%s = append(%s, ev)\n", indent, accessor, accessor))
					}
				}
				sb.WriteString(fmt.Sprintf("%s}\n", indent))
			} else {
				if isArray {
					sb.WriteString(fmt.Sprintf("%sfor len(fbz) > 0 && idx < %d {\n", indent, arrayLen))
				} else {
					sb.WriteString(fmt.Sprintf("%sfor len(fbz) > 0 {\n", indent))
				}
				if ert.Kind() == reflect.Ptr {
					sb.WriteString(fmt.Sprintf("%s\tvar ev %s\n", indent, ctx.goTypeName(ert.Elem())))
					ctx.writePrimitiveDecodeFrom(sb, "ev", einfo, fopts, indent+"\t", "fbz")
					if isArray {
						sb.WriteString(fmt.Sprintf("%s\t%s[idx] = &ev\n%s\tidx++\n", indent, accessor, indent))
					} else {
						sb.WriteString(fmt.Sprintf("%s\t%s = append(%s, &ev)\n", indent, accessor, accessor))
					}
				} else {
					sb.WriteString(fmt.Sprintf("%s\tvar ev %s\n", indent, ctx.goTypeName(ert)))
					ctx.writePrimitiveDecodeFrom(sb, "ev", einfo, fopts, indent+"\t", "fbz")
					if isArray {
						sb.WriteString(fmt.Sprintf("%s\t%s[idx] = ev\n%s\tidx++\n", indent, accessor, indent))
					} else {
						sb.WriteString(fmt.Sprintf("%s\t%s = append(%s, ev)\n", indent, accessor, accessor))
					}
				}
				sb.WriteString(fmt.Sprintf("%s}\n", indent))
			}
		}

	default:
		ctx.writePrimitiveDecode(sb, accessor, finfo, fopts, indent)
	}
}

func (ctx *P3Context2) writePointerFieldUnmarshal(sb *strings.Builder, accessor string, ftype reflect.Type, finfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	elemType := ftype.Elem()
	sb.WriteString(fmt.Sprintf("%s{\n", indent))
	sb.WriteString(fmt.Sprintf("%s\tvar pv %s\n", indent, ctx.goTypeName(elemType)))

	// Decode into pv using the field unmarshal logic.
	ctx.writeFieldUnmarshal(sb, "pv", finfo, fopts, indent+"\t")

	sb.WriteString(fmt.Sprintf("%s\t%s = &pv\n", indent, accessor))
	sb.WriteString(fmt.Sprintf("%s}\n", indent))
}

// === List / Repeated Field Decoding ===

func (ctx *P3Context2) writeUnpackedListUnmarshal(sb *strings.Builder, accessor string, finfo *amino.TypeInfo, fopts amino.FieldOptions, indent string, isArray bool, fieldName string) {
	ert := finfo.Type.Elem()
	einfo := finfo.Elem
	if einfo == nil {
		sb.WriteString(fmt.Sprintf("%s// TODO: nil Elem info for list type\n", indent))
		return
	}

	// beOptionByte: when element repr is uint8, amino encodes each element
	// as a raw byte rather than a varint (packed as a byte string).
	beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8
	typ3 := einfo.GetTyp3(fopts)

	arrayLen := 0
	if isArray {
		arrayLen = finfo.Type.Len()
	}

	// Helper to emit the "store element" line.
	storeElem := func(elemExpr string) string {
		if isArray {
			return fmt.Sprintf("if %s_idx >= %d {\n%s\treturn errors.New(\"array index out of bounds\")\n%s}\n%s%s[%s_idx] = %s\n%s%s_idx++\n",
				fieldName, arrayLen, indent, indent, indent, accessor, fieldName, elemExpr, indent, fieldName)
		}
		return fmt.Sprintf("%s = append(%s, %s)\n", accessor, accessor, elemExpr)
	}

	if typ3 != amino.Typ3ByteLength || beOptionByte {
		// Packed: decode ByteLength field, then decode elements from it.
		sb.WriteString(fmt.Sprintf("%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
		if beOptionByte {
			// Byte list: each element is a single raw byte.
			sb.WriteString(fmt.Sprintf("%sfor _, b := range fbz {\n", indent))
			if isArray {
				sb.WriteString(fmt.Sprintf("%s\tif %s_idx >= %d {\n%s\t\tbreak\n%s\t}\n", indent, fieldName, arrayLen, indent, indent))
			}
			if ert.Kind() == reflect.Ptr {
				sb.WriteString(fmt.Sprintf("%s\tev := %s(b)\n", indent, ctx.goTypeName(ert.Elem())))
				sb.WriteString(fmt.Sprintf("%s\t%s", indent, storeElem("&ev")))
			} else {
				sb.WriteString(fmt.Sprintf("%s\tev := %s(b)\n", indent, ctx.goTypeName(ert)))
				sb.WriteString(fmt.Sprintf("%s\t%s", indent, storeElem("ev")))
			}
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
		} else {
			if isArray {
				sb.WriteString(fmt.Sprintf("%sfor len(fbz) > 0 && %s_idx < %d {\n", indent, fieldName, arrayLen))
			} else {
				sb.WriteString(fmt.Sprintf("%sfor len(fbz) > 0 {\n", indent))
			}
			if ert.Kind() == reflect.Ptr {
				sb.WriteString(fmt.Sprintf("%s\tvar ev %s\n", indent, ctx.goTypeName(ert.Elem())))
				ctx.writePrimitiveDecodeFrom(sb, "ev", einfo, fopts, indent+"\t", "fbz")
				sb.WriteString(fmt.Sprintf("%s\tevp := &ev\n", indent))
				sb.WriteString(fmt.Sprintf("%s\t%s", indent, storeElem("evp")))
			} else {
				sb.WriteString(fmt.Sprintf("%s\tvar ev %s\n", indent, ctx.goTypeName(ert)))
				ctx.writePrimitiveDecodeFrom(sb, "ev", einfo, fopts, indent+"\t", "fbz")
				sb.WriteString(fmt.Sprintf("%s\t%s", indent, storeElem("ev")))
			}
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
		}
	} else {
		// Unpacked: this single field entry contains one element.
		ertIsPointer := ert.Kind() == reflect.Ptr
		writeImplicit := isListType(einfo.Type) &&
			einfo.Elem != nil &&
			einfo.Elem.ReprType.Type.Kind() != reflect.Uint8 &&
			einfo.Elem.ReprType.GetTyp3(fopts) != amino.Typ3ByteLength

		// Decode one element.
		if einfo.ReprType.Type.Kind() == reflect.Interface {
			// Interface element: decode via UnmarshalAny.
			sb.WriteString(fmt.Sprintf("%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent))
			sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
			sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
			sb.WriteString(fmt.Sprintf("%sif len(fbz) > 0 {\n", indent))
			sb.WriteString(fmt.Sprintf("%s\tvar ev %s\n", indent, ctx.goTypeName(ert)))
			sb.WriteString(fmt.Sprintf("%s\tif err := cdc.UnmarshalAny(fbz, &ev); err != nil {\n%s\t\treturn err\n%s\t}\n",
				indent, indent, indent))
			sb.WriteString(fmt.Sprintf("%s\t%s", indent, storeElem("ev")))
			sb.WriteString(fmt.Sprintf("%s} else {\n", indent))
			sb.WriteString(fmt.Sprintf("%s\t%s", indent, storeElem("nil")))
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
		} else if ertIsPointer {
			// Amino never encodes nil pointers in unpacked lists — nil elements are
			// skipped entirely. So 0x00 is always a valid length-prefix (empty message),
			// not a nil marker. Just decode the element unconditionally.
			sb.WriteString(fmt.Sprintf("%svar ev %s\n", indent, ctx.goTypeName(ert.Elem())))
			if writeImplicit {
				ctx.writeImplicitStructDecode(sb, "ev", einfo, fopts, indent)
			} else {
				ctx.writeByteSliceElementDecode(sb, "ev", einfo, fopts, indent)
			}
			sb.WriteString(fmt.Sprintf("%s%s", indent, storeElem("&ev")))
		} else {
			sb.WriteString(fmt.Sprintf("%svar ev %s\n", indent, ctx.goTypeName(ert)))
			if writeImplicit {
				ctx.writeImplicitStructDecode(sb, "ev", einfo, fopts, indent)
			} else {
				ctx.writeByteSliceElementDecode(sb, "ev", einfo, fopts, indent)
			}
			sb.WriteString(fmt.Sprintf("%s%s", indent, storeElem("ev")))
		}
	}
}

func (ctx *P3Context2) writeByteSliceElementDecode(sb *strings.Builder, accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	rinfo := einfo.ReprType
	rt := rinfo.Type

	isNestedList := isListType(rt) && rt.Elem().Kind() != reflect.Uint8
	if rt.Kind() == reflect.Struct || rt == reflect.TypeOf(time.Duration(0)) || isNestedList {
		// Struct-like elements (including time.Time, time.Duration, nested lists):
		// read length-prefixed data, then decode from fbz.
		sb.WriteString(fmt.Sprintf("%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))

		switch {
		case rt == reflect.TypeOf(time.Time{}):
			sb.WriteString(fmt.Sprintf("%s%s, _, err = amino.DecodeTime(fbz)\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		case rt == reflect.TypeOf(time.Duration(0)):
			sb.WriteString(fmt.Sprintf("%s%s, _, err = amino.DecodeDuration(fbz)\n", indent, accessor))
			sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		case isNestedList:
			sb.WriteString(fmt.Sprintf("%sif err := cdc.Unmarshal(fbz, &%s); err != nil {\n%s\treturn err\n%s}\n",
				indent, accessor, indent, indent))
		default:
			sb.WriteString(fmt.Sprintf("%sif err := %s.UnmarshalBinary2(cdc, fbz); err != nil {\n%s\treturn err\n%s}\n",
				indent, accessor, indent, indent))
		}
	} else {
		// Non-struct ByteLength element (string, []byte):
		// decode directly from bz (decode functions consume own length prefix).
		ctx.writePrimitiveDecodeFrom(sb, accessor, einfo, fopts, indent, "bz")
	}
}

func (ctx *P3Context2) writeImplicitStructDecode(sb *strings.Builder, accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	// Read outer ByteSlice (implicit struct).
	sb.WriteString(fmt.Sprintf("%sibz, n, err := amino.DecodeByteSlice(bz)\n", indent))
	sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
	sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
	sb.WriteString(fmt.Sprintf("%sif len(ibz) > 0 {\n", indent))
	// Read field 1 key from ibz.
	sb.WriteString(fmt.Sprintf("%s\t_, _, _n, _err := amino.DecodeFieldNumberAndTyp3(ibz)\n", indent))
	sb.WriteString(fmt.Sprintf("%s\tif _err != nil {\n%s\t\treturn _err\n%s\t}\n", indent, indent, indent))
	sb.WriteString(fmt.Sprintf("%s\tibz = ibz[_n:]\n", indent))
	// Read inner ByteSlice (packed data).
	sb.WriteString(fmt.Sprintf("%s\tfbz, _, _err2 := amino.DecodeByteSlice(ibz)\n", indent))
	sb.WriteString(fmt.Sprintf("%s\tif _err2 != nil {\n%s\t\treturn _err2\n%s\t}\n", indent, indent, indent))
	// Decode elements from packed data.
	innerEinfo := einfo.Elem
	if einfo.Type.Kind() == reflect.Array {
		length := einfo.Type.Len()
		for i := 0; i < length; i++ {
			elemAccessor := fmt.Sprintf("%s[%d]", accessor, i)
			sb.WriteString(fmt.Sprintf("%s\t{\n", indent))
			ctx.writePrimitiveDecodeFrom(sb, elemAccessor, innerEinfo, fopts, indent+"\t\t", "fbz")
			sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
		}
	} else {
		// Slice: decode while data remains.
		ert := einfo.Type.Elem()
		sb.WriteString(fmt.Sprintf("%s\tfor len(fbz) > 0 {\n", indent))
		sb.WriteString(fmt.Sprintf("%s\t\tvar _elem %s\n", indent, ctx.goTypeName(ert)))
		ctx.writePrimitiveDecodeFrom(sb, "_elem", innerEinfo, fopts, indent+"\t\t", "fbz")
		sb.WriteString(fmt.Sprintf("%s\t\t%s = append(%s, _elem)\n", indent, accessor, accessor))
		sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
	}
	sb.WriteString(fmt.Sprintf("%s}\n", indent))
}

func (ctx *P3Context2) writeInterfaceFieldUnmarshal(sb *strings.Builder, accessor string, indent string) {
	// Decode google.protobuf.Any: read ByteLength, decode typeURL + value.
	sb.WriteString(fmt.Sprintf("%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent))
	sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
	sb.WriteString(fmt.Sprintf("%sbz = bz[n:]\n", indent))
	sb.WriteString(fmt.Sprintf("%sif len(fbz) > 0 {\n", indent))
	sb.WriteString(fmt.Sprintf("%s\tif err := cdc.UnmarshalAny(fbz, &%s); err != nil {\n%s\t\treturn err\n%s\t}\n",
		indent, accessor, indent, indent))
	sb.WriteString(fmt.Sprintf("%s}\n", indent))
}

// === Primitive Decoding ===

func (ctx *P3Context2) writePrimitiveDecode(sb *strings.Builder, accessor string, info *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	ctx.writePrimitiveDecodeFrom(sb, accessor, info, fopts, indent, "bz")
}

func (ctx *P3Context2) writePrimitiveDecodeFrom(sb *strings.Builder, accessor string, info *amino.TypeInfo, fopts amino.FieldOptions, indent, srcVar string) {
	rinfo := info.ReprType
	rt := rinfo.Type
	kind := rt.Kind()

	switch kind {
	case reflect.Bool:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeBool(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Int8:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeVarint8(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Int16:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeVarint16(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Int32:
		if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeInt32(%s)\n", indent, srcVar))
		} else {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeVarint(%s)\n", indent, srcVar))
		}
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Int64:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeInt64(%s)\n", indent, srcVar))
		} else {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeVarint(%s)\n", indent, srcVar))
		}
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Int:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeVarint(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Uint8:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeUvarint8(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Uint16:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeUvarint16(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Uint32:
		if fopts.BinFixed32 {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeUint32(%s)\n", indent, srcVar))
		} else {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeUvarint(%s)\n", indent, srcVar))
		}
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Uint64:
		if fopts.BinFixed64 {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeUint64(%s)\n", indent, srcVar))
		} else {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeUvarint(%s)\n", indent, srcVar))
		}
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Uint:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeUvarint(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Float32:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeFloat32(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Float64:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeFloat64(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.String:
		sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeString(%s)\n", indent, srcVar))
		sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
		sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
		sb.WriteString(fmt.Sprintf("%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt)))

	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeByteSlice(%s)\n", indent, srcVar))
			sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
			sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
			sb.WriteString(fmt.Sprintf("%sif len(v) == 0 {\n%s\t%s = nil\n%s} else {\n%s\t%s = v\n%s}\n",
				indent, indent, accessor, indent, indent, accessor, indent))
		} else {
			sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive decode slice element kind %v\n", indent, rt.Elem().Kind()))
		}

	case reflect.Array:
		if rt.Elem().Kind() == reflect.Uint8 {
			sb.WriteString(fmt.Sprintf("%sv, n, err := amino.DecodeByteSlice(%s)\n", indent, srcVar))
			sb.WriteString(fmt.Sprintf("%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent))
			sb.WriteString(fmt.Sprintf("%s%s = %s[n:]\n", indent, srcVar, srcVar))
			sb.WriteString(fmt.Sprintf("%scopy(%s[:], v)\n", indent, accessor))
		} else {
			sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive decode array element kind %v\n", indent, rt.Elem().Kind()))
		}

	default:
		sb.WriteString(fmt.Sprintf("%s// TODO: unsupported primitive decode kind %v\n", indent, kind))
	}
}

// === Helpers ===

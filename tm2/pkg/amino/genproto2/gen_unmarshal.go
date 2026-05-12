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
	fmt.Fprintf(sb, "func (goo *%s) UnmarshalBinary2(cdc *amino.Codec, bz []byte, anyDepth int) error {\n", tname)

	if info.IsAminoMarshaler {
		rinfo := info.ReprType
		if err := ctx.writeReprUnmarshal(sb, rinfo); err != nil {
			return err
		}
		sb.WriteString("\treturn goo.UnmarshalAmino(repr)\n")
		sb.WriteString("}\n\n")
		return nil
	}

	// Handle struct types.
	if info.Type.Kind() == reflect.Struct {
		ctx.writeStructUnmarshalBody(sb, info, "goo")
		sb.WriteString("\treturn nil\n")
		sb.WriteString("}\n\n")
		return nil
	}

	// Handle non-struct primitive types (e.g. `type StringValue string`).
	// Decoded as implicit struct with a single field number 1.
	rt := info.Type

	// For array types, fall back to cdc.Unmarshal (the reflect path)
	// since writeReprUnmarshal uses append which doesn't work on arrays.
	if rt.Kind() == reflect.Array {
		sb.WriteString("\treturn cdc.UnmarshalReflect(bz, goo)\n")
		sb.WriteString("}\n\n")
		return nil
	}

	// Non-array, non-struct types: use writeReprUnmarshal which declares `var repr`.
	if err := ctx.writeReprUnmarshal(sb, info); err != nil {
		return err
	}
	fmt.Fprintf(sb, "\t*goo = %s(repr)\n", typeName(info))
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
			fmt.Fprintf(sb, "\tvar repr %s\n", ctx.goTypeName(rt))
			if rinfo.Registered {
				sb.WriteString("\tif err := repr.UnmarshalBinary2(cdc, bz, anyDepth); err != nil {\n\t\treturn err\n\t}\n")
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
		fmt.Fprintf(sb, "\tvar repr %s\n", ctx.goTypeName(rt))
		sb.WriteString("\tif len(bz) > 0 {\n")
		sb.WriteString("\t\tfnum, typ3, n, err := amino.DecodeFieldNumberAndTyp3(bz)\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		sb.WriteString("\t\tif fnum != 1 || typ3 != amino.Typ3ByteLength {\n")
		sb.WriteString("\t\t\treturn fmt.Errorf(\"repr field 1: expected ByteLength, got num=%v typ=%v\", fnum, typ3)\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tbz = bz[n:]\n")
		sb.WriteString("\t\tfbz, _, err := amino.DecodeByteSlice(bz)\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		ert := rt.Elem()
		beOptionByte := einfo.ReprType.Type.Kind() == reflect.Uint8
		if beOptionByte {
			// Each element is a raw byte.
			sb.WriteString("\t\tfor _, b := range fbz {\n")
			if einfo.IsAminoMarshaler {
				// For pointer element type, declare elem as the value type to
				// give UnmarshalAmino a valid receiver, then append &elem.
				elemType := ert
				if ert.Kind() == reflect.Ptr {
					elemType = ert.Elem()
				}
				fmt.Fprintf(sb, "\t\t\tvar elem %s\n", ctx.goTypeName(elemType))
				sb.WriteString("\t\t\tif err := elem.UnmarshalAmino(b); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
				if ert.Kind() == reflect.Ptr {
					sb.WriteString("\t\t\trepr = append(repr, &elem)\n")
				} else {
					sb.WriteString("\t\t\trepr = append(repr, elem)\n")
				}
			} else {
				fmt.Fprintf(sb, "\t\t\telem := %s(b)\n", ctx.goTypeName(ert))
				sb.WriteString("\t\t\trepr = append(repr, elem)\n")
			}
			sb.WriteString("\t\t}\n")
		} else {
			sb.WriteString("\t\tfor len(fbz) > 0 {\n")
			if einfo.IsAminoMarshaler {
				reprType := einfo.ReprType.Type
				fmt.Fprintf(sb, "\t\t\tvar rv %s\n", ctx.goTypeName(reprType))
				ctx.writePrimitiveDecodeFrom(sb, "rv", einfo.ReprType, fopts, "\t\t\t", "fbz")
				// Pointer element: declare elem as value type for UnmarshalAmino.
				elemType := ert
				if ert.Kind() == reflect.Ptr {
					elemType = ert.Elem()
				}
				fmt.Fprintf(sb, "\t\t\tvar elem %s\n", ctx.goTypeName(elemType))
				sb.WriteString("\t\t\tif err := elem.UnmarshalAmino(rv); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
				if ert.Kind() == reflect.Ptr {
					sb.WriteString("\t\t\trepr = append(repr, &elem)\n")
				} else {
					sb.WriteString("\t\t\trepr = append(repr, elem)\n")
				}
			} else {
				fmt.Fprintf(sb, "\t\t\tvar elem %s\n", ctx.goTypeName(ert))
				ctx.writePrimitiveDecodeFrom(sb, "elem", einfo, fopts, "\t\t\t", "fbz")
				sb.WriteString("\t\t\trepr = append(repr, elem)\n")
			}
			sb.WriteString("\t\t}\n")
		}
		sb.WriteString("\t}\n")
	} else {
		// Primitive repr: read field 1 key, then decode value.
		fmt.Fprintf(sb, "\tvar repr %s\n", ctx.goTypeName(rt))
		sb.WriteString("\tif len(bz) > 0 {\n")
		expectedTyp3 := rinfo.GetTyp3(fopts)
		sb.WriteString("\t\tfnum, typ3, n, err := amino.DecodeFieldNumberAndTyp3(bz)\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		fmt.Fprintf(sb, "\t\tif fnum != 1 || typ3 != %s {\n", typ3GoStr(expectedTyp3))
		fmt.Fprintf(sb, "\t\t\treturn fmt.Errorf(\"repr field 1: expected typ3 %%v, got num=%%v typ=%%v\", %s, fnum, typ3)\n", typ3GoStr(expectedTyp3))
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tbz = bz[n:]\n")

		switch rt.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sb.WriteString("\t\tv, _, err := amino.DecodeVarint(bz)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			fmt.Fprintf(sb, "\t\trepr = %s(v)\n", ctx.goTypeName(rt))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			sb.WriteString("\t\tv, _, err := amino.DecodeUvarint(bz)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			fmt.Fprintf(sb, "\t\trepr = %s(v)\n", ctx.goTypeName(rt))
		case reflect.String:
			sb.WriteString("\t\tv, _, err := amino.DecodeString(bz)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			fmt.Fprintf(sb, "\t\trepr = %s(v)\n", ctx.goTypeName(rt))
		case reflect.Bool:
			sb.WriteString("\t\tv, _, err := amino.DecodeBool(bz)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
			fmt.Fprintf(sb, "\t\trepr = %s(v)\n", ctx.goTypeName(rt))
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
	fmt.Fprintf(sb, "\tvar repr %s\n", ctx.goTypeName(info.Type))

	if typ3 != amino.Typ3ByteLength {
		// Packed: whole bz is the packed data.
		sb.WriteString("\tfor len(bz) > 0 {\n")
		if einfo.IsAminoMarshaler {
			// Decode repr type, then UnmarshalAmino.
			reprType := einfo.ReprType.Type
			fmt.Fprintf(sb, "\t\tvar rv %s\n", ctx.goTypeName(reprType))
			ctx.writePrimitiveDecode(sb, "rv", einfo.ReprType, fopts, "\t\t")
			fmt.Fprintf(sb, "\t\tvar elem %s\n", ctx.goTypeName(ert))
			sb.WriteString("\t\tif err := elem.UnmarshalAmino(rv); err != nil {\n\t\t\treturn err\n\t\t}\n")
		} else {
			fmt.Fprintf(sb, "\t\tvar elem %s\n", ctx.goTypeName(ert))
			ctx.writePrimitiveDecode(sb, "elem", einfo, fopts, "\t\t")
		}
		sb.WriteString("\t\trepr = append(repr, elem)\n")
		sb.WriteString("\t}\n")
	} else {
		// Unpacked: repeated field entries with field number 1, ByteLength typ3.
		sb.WriteString("\tfor len(bz) > 0 {\n")
		sb.WriteString("\t\tfnum, typ3, n, err := amino.DecodeFieldNumberAndTyp3(bz)\n")
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
		sb.WriteString("\t\tbz = bz[n:]\n")
		sb.WriteString("\t\tif fnum != 1 {\n")
		sb.WriteString("\t\t\treturn fmt.Errorf(\"unknown field number %d in unpacked slice repr (expected 1)\", fnum)\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tif typ3 != amino.Typ3ByteLength {\n")
		sb.WriteString("\t\t\treturn fmt.Errorf(\"unpacked slice repr: expected field 1 ByteLength, got typ=%v\", typ3)\n")
		sb.WriteString("\t\t}\n")
		fmt.Fprintf(sb, "\t\tvar elem %s\n", ctx.goTypeName(ert))
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
			fmt.Fprintf(sb, "\tvar %s_idx int\n", field.Name)
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
			fmt.Fprintf(sb, "\t%s.%s = time.Unix(0, 0).UTC()\n", recv, field.Name)
		}
	}

	sb.WriteString("\tvar lastFieldNum uint32\n")
	sb.WriteString("\tfor len(bz) > 0 {\n")
	sb.WriteString("\t\tfnum, typ3, n, err := amino.DecodeFieldNumberAndTyp3(bz)\n")
	sb.WriteString("\t\t_ = typ3\n")
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
		fmt.Fprintf(sb, "\t\tcase %d:\n", fnum)

		// Validate that the received wire type matches what this field expects.
		// Without this check, malformed wire data would be silently misparsed.
		expectedTyp3 := finfo.GetTyp3(fopts)
		fmt.Fprintf(sb, "\t\t\tif typ3 != %s {\n", typ3GoStr(expectedTyp3))
		fmt.Fprintf(sb, "\t\t\t\treturn fmt.Errorf(\"field %d: expected typ3 %%v, got %%v\", %s, typ3)\n", fnum, typ3GoStr(expectedTyp3))
		sb.WriteString("\t\t\t}\n")

		if field.UnpackedList {
			isArray := ftype.Kind() == reflect.Array
			ctx.writeUnpackedListUnmarshal(sb, accessor, finfo, fopts, "\t\t\t", isArray, fname)
			fmt.Fprintf( // Continue consuming repeated entries with the same field number.
				sb, "\t\t\tfor len(bz) > 0 {\n")
			fmt.Fprintf(sb, "\t\t\t\tvar nextFnum uint32\n")
			fmt.Fprintf(sb, "\t\t\t\tvar nextTyp3 amino.Typ3\n")
			fmt.Fprintf(sb, "\t\t\t\tnextFnum, nextTyp3, n, err = amino.DecodeFieldNumberAndTyp3(bz)\n")
			fmt.Fprintf(sb, "\t\t\t\tif err != nil {\n\t\t\t\t\treturn err\n\t\t\t\t}\n")
			fmt.Fprintf(sb, "\t\t\t\tif nextFnum != %d {\n\t\t\t\t\tbreak\n\t\t\t\t}\n", fnum)
			fmt.Fprintf(sb, "\t\t\t\tif nextTyp3 != %s {\n\t\t\t\t\treturn fmt.Errorf(\"field %d: expected typ3 %%v, got %%v\", %s, nextTyp3)\n\t\t\t\t}\n", typ3GoStr(expectedTyp3), fnum, typ3GoStr(expectedTyp3))
			fmt.Fprintf(sb, "\t\t\t\tbz = bz[n:]\n")
			ctx.writeUnpackedListUnmarshal(sb, accessor, finfo, fopts, "\t\t\t\t", isArray, fname)
			fmt.Fprintf(sb, "\t\t\t}\n")
		} else if isPtr {
			ctx.writePointerFieldUnmarshal(sb, accessor, ftype, finfo, fopts, "\t\t\t")
		} else if finfo.IsAminoMarshaler {
			// AminoMarshaler field: decode repr, then UnmarshalAmino.
			reprTypeName := ctx.goTypeName(finfo.ReprType.Type)
			fmt.Fprintf(sb, "\t\t\tvar repr %s\n", reprTypeName)
			ctx.writeFieldUnmarshal(sb, "repr", finfo.ReprType, fopts, "\t\t\t")
			fmt.Fprintf(sb, "\t\t\tif err := %s.UnmarshalAmino(repr); err != nil {\n\t\t\t\treturn err\n\t\t\t}\n", accessor)
		} else {
			ctx.writeFieldUnmarshal(sb, accessor, finfo, fopts, "\t\t\t")
		}
	}

	sb.WriteString("\t\tdefault:\n")
	fmt.Fprintf(sb, "\t\t\treturn fmt.Errorf(\"unknown field number %%d for %s\", fnum)\n", info.Type.Name())
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
			fmt.Fprintf(sb, "\tif %s == nil {\n\t\tv := time.Unix(0, 0).UTC()\n\t\t%s = &v\n\t}\n", accessor, accessor)
		} else if ert.Kind() != reflect.Struct {
			fmt.Fprintf(sb, "\tif %s == nil {\n\t\t%s = new(%s)\n\t}\n", accessor, accessor, ctx.goTypeName(ert))
		}
	}
}

func (ctx *P3Context2) writeFieldUnmarshal(sb *strings.Builder, accessor string, finfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	rinfo := finfo.ReprType
	rt := rinfo.Type

	switch {
	case rt == reflect.TypeOf(time.Time{}):
		fmt.Fprintf(sb, "%s// time.Time (ByteLength)\n", indent)
		fmt.Fprintf(sb, "%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
		fmt.Fprintf(sb, "%s%s, _, err = amino.DecodeTime(fbz)\n", indent, accessor)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)

	case rt == reflect.TypeOf(time.Duration(0)):
		fmt.Fprintf(sb, "%s// time.Duration (ByteLength)\n", indent)
		fmt.Fprintf(sb, "%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
		fmt.Fprintf(sb, "%s%s, _, err = amino.DecodeDuration(fbz)\n", indent, accessor)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)

	case rt.Kind() == reflect.Struct:
		fmt.Fprintf(sb, "%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
		fmt.Fprintf(sb, "%sif err := %s.UnmarshalBinary2(cdc, fbz, anyDepth); err != nil {\n%s\treturn err\n%s}\n",
			indent, accessor, indent, indent)

	case rt.Kind() == reflect.Interface:
		ctx.writeInterfaceFieldUnmarshal(sb, accessor, indent)

	case rt.Kind() == reflect.String:
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeString(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case rt.Kind() == reflect.Slice && rt.Elem().Kind() == reflect.Uint8:
		// []byte
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeByteSlice(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
		fmt.Fprintf(sb, "%sif len(v) == 0 {\n%s\t%s = nil\n%s} else {\n%s\t%s = v\n%s}\n",
			indent, indent, accessor, indent, indent, accessor, indent)

	case rt.Kind() == reflect.Array && rt.Elem().Kind() == reflect.Uint8:
		// [N]byte — read length prefix then copy directly from buffer (no intermediate slice).
		arrLen := rt.Len()
		fmt.Fprintf(sb, "%svar count uint64\n", indent)
		fmt.Fprintf(sb, "%scount, n, err = amino.DecodeUvarint(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
		fmt.Fprintf(sb, "%sif int(count) != %d {\n", indent, arrLen)
		fmt.Fprintf(sb, "%s\treturn fmt.Errorf(\"invalid [%d]byte length: expected %d, got %%d\", count)\n", indent, arrLen, arrLen)
		fmt.Fprintf(sb, "%s}\n", indent)
		fmt.Fprintf(sb, "%sif len(bz) < %d {\n", indent, arrLen)
		fmt.Fprintf(sb, "%s\treturn fmt.Errorf(\"insufficient bytes for [%d]byte: have %%d\", len(bz))\n", indent, arrLen)
		fmt.Fprintf(sb, "%s}\n", indent)
		fmt.Fprintf(sb, "%scopy(%s[:], bz[:%d])\n", indent, accessor, arrLen)
		fmt.Fprintf(sb, "%sbz = bz[%d:]\n", indent, arrLen)

	case isListType(rt) && rt.Elem().Kind() != reflect.Uint8:
		// Packed list (non-byte elements): decode length-prefixed block, then elements.
		fmt.Fprintf(sb, "%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
		einfo := finfo.Elem
		ert := rt.Elem()
		if einfo != nil {
			isArray := rt.Kind() == reflect.Array
			arrayLen := 0
			if isArray {
				arrayLen = rt.Len()
				fmt.Fprintf(sb, "%svar idx int\n", indent)
			}
			if einfo.ReprType.Type.Kind() == reflect.Uint8 {
				fmt.Fprintf( // List of (repr) bytes: elements are raw bytes (beOptionByte encoding).
					sb, "%sfor _, b := range fbz {\n", indent)
				if isArray {
					fmt.Fprintf(sb, "%s\tif idx >= %d {\n%s\t\tbreak\n%s\t}\n", indent, arrayLen, indent, indent)
				}
				if einfo.IsAminoMarshaler {
					// For pointer elements, declare elem as the value type so
					// UnmarshalAmino has a valid (non-nil) receiver; later we
					// append &elem to the pointer-slice accessor.
					elemType := ert
					if ert.Kind() == reflect.Ptr {
						elemType = ert.Elem()
					}
					fmt.Fprintf(sb, "%s\tvar elem %s\n", indent, ctx.goTypeName(elemType))
					fmt.Fprintf(sb, "%s\tif err := elem.UnmarshalAmino(b); err != nil {\n%s\t\treturn err\n%s\t}\n", indent, indent, indent)
				} else if ert.Kind() == reflect.Ptr {
					fmt.Fprintf(sb, "%s\tev := %s(b)\n", indent, ctx.goTypeName(ert.Elem()))
				} else {
					fmt.Fprintf(sb, "%s\tev := %s(b)\n", indent, ctx.goTypeName(ert))
				}
				if einfo.IsAminoMarshaler {
					if ert.Kind() == reflect.Ptr {
						// Slice is []*T; store &elem.
						if isArray {
							fmt.Fprintf(sb, "%s\t%s[idx] = &elem\n%s\tidx++\n", indent, accessor, indent)
						} else {
							fmt.Fprintf(sb, "%s\t%s = append(%s, &elem)\n", indent, accessor, accessor)
						}
					} else {
						if isArray {
							fmt.Fprintf(sb, "%s\t%s[idx] = elem\n%s\tidx++\n", indent, accessor, indent)
						} else {
							fmt.Fprintf(sb, "%s\t%s = append(%s, elem)\n", indent, accessor, accessor)
						}
					}
				} else if ert.Kind() == reflect.Ptr {
					if isArray {
						fmt.Fprintf(sb, "%s\t%s[idx] = &ev\n%s\tidx++\n", indent, accessor, indent)
					} else {
						fmt.Fprintf(sb, "%s\t%s = append(%s, &ev)\n", indent, accessor, accessor)
					}
				} else {
					if isArray {
						fmt.Fprintf(sb, "%s\t%s[idx] = ev\n%s\tidx++\n", indent, accessor, indent)
					} else {
						fmt.Fprintf(sb, "%s\t%s = append(%s, ev)\n", indent, accessor, accessor)
					}
				}
				fmt.Fprintf(sb, "%s}\n", indent)
			} else {
				if isArray {
					fmt.Fprintf(sb, "%sfor len(fbz) > 0 && idx < %d {\n", indent, arrayLen)
				} else {
					fmt.Fprintf(sb, "%sfor len(fbz) > 0 {\n", indent)
				}
				if ert.Kind() == reflect.Ptr {
					fmt.Fprintf(sb, "%s\tvar ev %s\n", indent, ctx.goTypeName(ert.Elem()))
					ctx.writePrimitiveDecodeFrom(sb, "ev", einfo, fopts, indent+"\t", "fbz")
					if isArray {
						fmt.Fprintf(sb, "%s\t%s[idx] = &ev\n%s\tidx++\n", indent, accessor, indent)
					} else {
						fmt.Fprintf(sb, "%s\t%s = append(%s, &ev)\n", indent, accessor, accessor)
					}
				} else {
					fmt.Fprintf(sb, "%s\tvar ev %s\n", indent, ctx.goTypeName(ert))
					ctx.writePrimitiveDecodeFrom(sb, "ev", einfo, fopts, indent+"\t", "fbz")
					if isArray {
						fmt.Fprintf(sb, "%s\t%s[idx] = ev\n%s\tidx++\n", indent, accessor, indent)
					} else {
						fmt.Fprintf(sb, "%s\t%s = append(%s, ev)\n", indent, accessor, accessor)
					}
				}
				fmt.Fprintf(sb, "%s}\n", indent)
			}
		}

	default:
		ctx.writePrimitiveDecode(sb, accessor, finfo, fopts, indent)
	}
}

func (ctx *P3Context2) writePointerFieldUnmarshal(sb *strings.Builder, accessor string, ftype reflect.Type, finfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	elemType := ftype.Elem()
	fmt.Fprintf(sb, "%s{\n", indent)
	fmt.Fprintf(sb, "%s\tvar pv %s\n", indent, ctx.goTypeName(elemType))

	// Decode into pv using the field unmarshal logic.
	ctx.writeFieldUnmarshal(sb, "pv", finfo, fopts, indent+"\t")
	fmt.Fprintf(sb, "%s\t%s = &pv\n", indent, accessor)
	fmt.Fprintf(sb, "%s}\n", indent)
}

// === List / Repeated Field Decoding ===

func (ctx *P3Context2) writeUnpackedListUnmarshal(sb *strings.Builder, accessor string, finfo *amino.TypeInfo, fopts amino.FieldOptions, indent string, isArray bool, fieldName string) {
	ert := finfo.Type.Elem()
	einfo := finfo.Elem
	if einfo == nil {
		panic(fmt.Sprintf("genproto2: writeUnpackedListUnmarshal: nil Elem info for list type %v (accessor=%s)", finfo.Type, accessor))
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
		fmt.Fprintf(sb, "%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
		if beOptionByte {
			fmt.Fprintf( // Byte list: each element is a single raw byte.
				sb, "%sfor _, b := range fbz {\n", indent)
			if isArray {
				fmt.Fprintf(sb, "%s\tif %s_idx >= %d {\n%s\t\tbreak\n%s\t}\n", indent, fieldName, arrayLen, indent, indent)
			}
			if ert.Kind() == reflect.Ptr {
				fmt.Fprintf(sb, "%s\tev := %s(b)\n", indent, ctx.goTypeName(ert.Elem()))
				fmt.Fprintf(sb, "%s\t%s", indent, storeElem("&ev"))
			} else {
				fmt.Fprintf(sb, "%s\tev := %s(b)\n", indent, ctx.goTypeName(ert))
				fmt.Fprintf(sb, "%s\t%s", indent, storeElem("ev"))
			}
			fmt.Fprintf(sb, "%s}\n", indent)
		} else {
			if isArray {
				fmt.Fprintf(sb, "%sfor len(fbz) > 0 && %s_idx < %d {\n", indent, fieldName, arrayLen)
			} else {
				fmt.Fprintf(sb, "%sfor len(fbz) > 0 {\n", indent)
			}
			if ert.Kind() == reflect.Ptr {
				fmt.Fprintf(sb, "%s\tvar ev %s\n", indent, ctx.goTypeName(ert.Elem()))
				ctx.writePrimitiveDecodeFrom(sb, "ev", einfo, fopts, indent+"\t", "fbz")
				fmt.Fprintf(sb, "%s\tevp := &ev\n", indent)
				fmt.Fprintf(sb, "%s\t%s", indent, storeElem("evp"))
			} else {
				fmt.Fprintf(sb, "%s\tvar ev %s\n", indent, ctx.goTypeName(ert))
				ctx.writePrimitiveDecodeFrom(sb, "ev", einfo, fopts, indent+"\t", "fbz")
				fmt.Fprintf(sb, "%s\t%s", indent, storeElem("ev"))
			}
			fmt.Fprintf(sb, "%s}\n", indent)
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
			fmt.Fprintf( // Interface element: decode via UnmarshalAny.
				sb, "%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent)
			fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
			fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
			fmt.Fprintf(sb, "%sif len(fbz) > 0 {\n", indent)
			fmt.Fprintf(sb, "%s\tvar ev %s\n", indent, ctx.goTypeName(ert))
			fmt.Fprintf(sb, "%s\tif err := cdc.UnmarshalAnyBinary2(fbz, &ev, anyDepth); err != nil {\n%s\t\treturn err\n%s\t}\n",
				indent, indent, indent)
			fmt.Fprintf(sb, "%s\t%s", indent, storeElem("ev"))
			fmt.Fprintf(sb, "%s} else {\n", indent)
			fmt.Fprintf(sb, "%s\t%s", indent, storeElem("nil"))
			fmt.Fprintf(sb, "%s}\n", indent)
		} else if ertIsPointer {
			fmt.Fprintf( // Amino never encodes nil pointers in unpacked lists — nil elements are
				// skipped entirely. So 0x00 is always a valid length-prefix (empty message),
				// not a nil marker. Just decode the element unconditionally.
				sb, "%svar ev %s\n", indent, ctx.goTypeName(ert.Elem()))
			if writeImplicit {
				ctx.writeImplicitStructDecode(sb, "ev", einfo, fopts, indent)
			} else {
				ctx.writeByteSliceElementDecode(sb, "ev", einfo, fopts, indent)
			}
			fmt.Fprintf(sb, "%s%s", indent, storeElem("&ev"))
		} else {
			fmt.Fprintf(sb, "%svar ev %s\n", indent, ctx.goTypeName(ert))
			if writeImplicit {
				ctx.writeImplicitStructDecode(sb, "ev", einfo, fopts, indent)
			} else {
				ctx.writeByteSliceElementDecode(sb, "ev", einfo, fopts, indent)
			}
			fmt.Fprintf(sb, "%s%s", indent, storeElem("ev"))
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
		fmt.Fprintf(sb, "%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)

		switch {
		case rt == reflect.TypeOf(time.Time{}):
			fmt.Fprintf(sb, "%s%s, _, err = amino.DecodeTime(fbz)\n", indent, accessor)
			fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		case rt == reflect.TypeOf(time.Duration(0)):
			fmt.Fprintf(sb, "%s%s, _, err = amino.DecodeDuration(fbz)\n", indent, accessor)
			fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		case isNestedList:
			fmt.Fprintf(sb, "%sif err := cdc.Unmarshal(fbz, &%s); err != nil {\n%s\treturn err\n%s}\n",
				indent, accessor, indent, indent)
		default:
			fmt.Fprintf(sb, "%sif err := %s.UnmarshalBinary2(cdc, fbz, anyDepth); err != nil {\n%s\treturn err\n%s}\n",
				indent, accessor, indent, indent)
		}
	} else if einfo.IsAminoMarshaler {
		// AminoMarshaler element: decode repr, then UnmarshalAmino.
		reprType := einfo.ReprType.Type
		fmt.Fprintf(sb, "%svar rv %s\n", indent, ctx.goTypeName(reprType))
		ctx.writePrimitiveDecodeFrom(sb, "rv", einfo.ReprType, fopts, indent, "bz")
		fmt.Fprintf(sb, "%sif err := %s.UnmarshalAmino(rv); err != nil {\n%s\treturn err\n%s}\n",
			indent, accessor, indent, indent)
	} else {
		// Non-struct ByteLength element (string, []byte):
		// decode directly from bz (decode functions consume own length prefix).
		ctx.writePrimitiveDecodeFrom(sb, accessor, einfo, fopts, indent, "bz")
	}
}

func (ctx *P3Context2) writeImplicitStructDecode(sb *strings.Builder, accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	// Read outer ByteSlice (implicit struct).
	fmt.Fprintf(sb, "%sibz, n, err := amino.DecodeByteSlice(bz)\n", indent)
	fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
	fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
	fmt.Fprintf(sb, "%sif len(ibz) > 0 {\n", indent)
	// Read field 1 key from ibz and validate it's the expected ByteLength wrapper.
	fmt.Fprintf(sb, "%s\t_fnum, _typ3, _n, _err := amino.DecodeFieldNumberAndTyp3(ibz)\n", indent)
	fmt.Fprintf(sb, "%s\tif _err != nil {\n%s\t\treturn _err\n%s\t}\n", indent, indent, indent)
	fmt.Fprintf(sb, "%s\tif _fnum != 1 || _typ3 != amino.Typ3ByteLength {\n", indent)
	fmt.Fprintf(sb, "%s\t\treturn fmt.Errorf(\"implicit struct: expected field 1 ByteLength, got num=%%v typ=%%v\", _fnum, _typ3)\n", indent)
	fmt.Fprintf(sb, "%s\t}\n", indent)
	fmt.Fprintf(sb, "%s\tibz = ibz[_n:]\n", indent)
	// Read inner ByteSlice (packed data).
	fmt.Fprintf(sb, "%s\tfbz, _fbn, _err2 := amino.DecodeByteSlice(ibz)\n", indent)
	fmt.Fprintf(sb, "%s\tif _err2 != nil {\n%s\t\treturn _err2\n%s\t}\n", indent, indent, indent)
	// The implicit struct has only field 1 — reject anything past it.
	fmt.Fprintf(sb, "%s\tif len(ibz)-_fbn > 0 {\n", indent)
	fmt.Fprintf(sb, "%s\t\treturn fmt.Errorf(\"implicit struct: %%d trailing bytes after field 1\", len(ibz)-_fbn)\n", indent)
	fmt.Fprintf(sb, "%s\t}\n", indent)
	// Decode elements from packed data.
	innerEinfo := einfo.Elem
	if einfo.Type.Kind() == reflect.Array {
		length := einfo.Type.Len()
		for i := 0; i < length; i++ {
			elemAccessor := fmt.Sprintf("%s[%d]", accessor, i)
			fmt.Fprintf(sb, "%s\t{\n", indent)
			ctx.writePrimitiveDecodeFrom(sb, elemAccessor, innerEinfo, fopts, indent+"\t\t", "fbz")
			fmt.Fprintf(sb, "%s\t}\n", indent)
		}
	} else {
		// Slice: decode while data remains.
		ert := einfo.Type.Elem()
		fmt.Fprintf(sb, "%s\tfor len(fbz) > 0 {\n", indent)
		fmt.Fprintf(sb, "%s\t\tvar _elem %s\n", indent, ctx.goTypeName(ert))
		ctx.writePrimitiveDecodeFrom(sb, "_elem", innerEinfo, fopts, indent+"\t\t", "fbz")
		fmt.Fprintf(sb, "%s\t\t%s = append(%s, _elem)\n", indent, accessor, accessor)
		fmt.Fprintf(sb, "%s\t}\n", indent)
	}
	fmt.Fprintf(sb, "%s}\n", indent)
}

func (ctx *P3Context2) writeInterfaceFieldUnmarshal(sb *strings.Builder, accessor string, indent string) {
	// Decode google.protobuf.Any: read ByteLength, decode typeURL + value.
	fmt.Fprintf(sb, "%sfbz, n, err := amino.DecodeByteSlice(bz)\n", indent)
	fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
	fmt.Fprintf(sb, "%sbz = bz[n:]\n", indent)
	fmt.Fprintf(sb, "%sif len(fbz) > 0 {\n", indent)
	fmt.Fprintf(sb, "%s\tif err := cdc.UnmarshalAnyBinary2(fbz, &%s, anyDepth); err != nil {\n%s\t\treturn err\n%s\t}\n",
		indent, accessor, indent, indent)
	fmt.Fprintf(sb, "%s}\n", indent)
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
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeBool(%s)\n", indent, srcVar)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Int8:
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeVarint8(%s)\n", indent, srcVar)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Int16:
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeVarint16(%s)\n", indent, srcVar)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Int32:
		if fopts.BinFixed32 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeInt32(%s)\n", indent, srcVar)
		} else {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeVarint(%s)\n", indent, srcVar)
		}
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Int64:
		if fopts.BinFixed64 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeInt64(%s)\n", indent, srcVar)
		} else {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeVarint(%s)\n", indent, srcVar)
		}
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Int:
		if fopts.BinFixed64 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeInt64(%s)\n", indent, srcVar)
		} else if fopts.BinFixed32 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeInt32(%s)\n", indent, srcVar)
		} else {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeVarint(%s)\n", indent, srcVar)
		}
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Uint8:
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUvarint8(%s)\n", indent, srcVar)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Uint16:
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUvarint16(%s)\n", indent, srcVar)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Uint32:
		if fopts.BinFixed32 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUint32(%s)\n", indent, srcVar)
		} else {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUvarint(%s)\n", indent, srcVar)
		}
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Uint64:
		if fopts.BinFixed64 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUint64(%s)\n", indent, srcVar)
		} else {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUvarint(%s)\n", indent, srcVar)
		}
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Uint:
		if fopts.BinFixed64 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUint64(%s)\n", indent, srcVar)
		} else if fopts.BinFixed32 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUint32(%s)\n", indent, srcVar)
		} else {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeUvarint(%s)\n", indent, srcVar)
		}
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Float32:
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeFloat32(%s)\n", indent, srcVar)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Float64:
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeFloat64(%s)\n", indent, srcVar)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.String:
		fmt.Fprintf(sb, "%sv, n, err := amino.DecodeString(%s)\n", indent, srcVar)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
		fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
		fmt.Fprintf(sb, "%s%s = %s(v)\n", indent, accessor, ctx.goTypeName(rt))

	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeByteSlice(%s)\n", indent, srcVar)
			fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
			fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
			fmt.Fprintf(sb, "%sif len(v) == 0 {\n%s\t%s = nil\n%s} else {\n%s\t%s = v\n%s}\n",
				indent, indent, accessor, indent, indent, accessor, indent)
		} else {
			panic(fmt.Sprintf("genproto2: writePrimitiveDecodeFrom: unsupported slice element kind %v (type=%v, accessor=%s)", rt.Elem().Kind(), rt, accessor))
		}

	case reflect.Array:
		if rt.Elem().Kind() == reflect.Uint8 {
			fmt.Fprintf(sb, "%sv, n, err := amino.DecodeByteSlice(%s)\n", indent, srcVar)
			fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn err\n%s}\n", indent, indent, indent)
			fmt.Fprintf(sb, "%s%s = %s[n:]\n", indent, srcVar, srcVar)
			fmt.Fprintf(sb, "%scopy(%s[:], v)\n", indent, accessor)
		} else {
			panic(fmt.Sprintf("genproto2: writePrimitiveDecodeFrom: unsupported array element kind %v (type=%v, accessor=%s)", rt.Elem().Kind(), rt, accessor))
		}

	default:
		panic(fmt.Sprintf("genproto2: writePrimitiveDecodeFrom: unsupported kind %v (type=%v, accessor=%s)", kind, rt, accessor))
	}
}

// === Helpers ===

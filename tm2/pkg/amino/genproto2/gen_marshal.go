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
	fmt.Fprintf(sb, "func (goo %s) MarshalBinary2(cdc *amino.Codec, buf []byte, offset int) (int, error) {\n", tname)
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

	// Handle struct types.
	if info.Type.Kind() == reflect.Struct {
		if err := ctx.writeStructMarshalBody(sb, info, "goo"); err != nil {
			return err
		}
		sb.WriteString("\treturn offset, err\n")
		sb.WriteString("}\n\n")
		return nil
	}

	// Handle non-struct primitive types (e.g. `type StringValue string`).
	// Encoded as implicit struct with a single field number 1.
	sb.WriteString("\trepr := goo\n")
	if err := ctx.writeReprMarshal(sb, info); err != nil {
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
		if !rinfo.IsStructOrUnpackedTopLevel() {
			return ctx.writePackedSliceReprMarshal(sb, rinfo)
		}
		return ctx.writeSliceReprMarshal(sb, rinfo)

	default:
		// Primitive repr wrapped in implicit struct field 1.
		typ3 := rinfo.GetTyp3(fopts)
		sb.WriteString("\t{\n")
		sb.WriteString("\t\tbefore := offset\n")
		ctx.writePrimitiveEncode(sb, "repr", rinfo, fopts, "\t\t")
		// Match writeFieldIfNotEmpty at binary_encode.go:592-596: roll back
		// if the value encoded to nothing or to a single 0x00 byte. The
		// PrependXXX helpers write backward, so the single byte sits at
		// buf[offset] after encoding.
		sb.WriteString("\t\tvalueLen := before - offset\n")
		sb.WriteString("\t\tif valueLen > 1 || (valueLen == 1 && buf[offset] != 0x00) {\n")
		fmt.Fprintf(sb, "\t\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, %s)\n", typ3GoStr(typ3))
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
		fmt.Fprintf(sb, "\tif %s != nil {\n", accessor)
		derefAccessor := fmt.Sprintf("(*%s)", accessor)
		if !field.WriteEmpty {
			zeroCheck := ctx.zeroCheck(derefAccessor, finfo, fopts)
			if zeroCheck != "" {
				fmt.Fprintf(sb, "\t\tif %s {\n", zeroCheck)
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
	// The "should I emit this field?" decision must be based on the REPR
	// value's zeroness, not the original Go value's. Example: crypto.Address
	// is [20]byte but MarshalAmino returns a bech32 string — the zero
	// Address produces a non-empty string ("g1qqq...luuxe"), so the field
	// MUST be emitted. Checking the original [20]byte's zeroness (the old
	// code path) would incorrectly omit it, diverging from the reflect
	// encoder which rolls back only when the serialized value is a single
	// 0x00 byte (see binary_encode.go writeFieldIfNotEmpty).
	if finfo.IsAminoMarshaler && finfo.Type.Kind() != reflect.Struct {
		fmt.Fprintf(sb, "\t{\n")
		fmt.Fprintf(sb, "\t\trepr, err := %s.MarshalAmino()\n", accessor)
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
		reprZeroCheck := ctx.zeroCheck("repr", finfo.ReprType, fopts)
		if reprZeroCheck != "" && !field.WriteEmpty {
			fmt.Fprintf(sb, "\t\tif %s {\n", reprZeroCheck)
			ctx.writeFieldValueMarshal(sb, "repr", fnum, finfo.ReprType, fopts, false, "\t\t\t")
			sb.WriteString("\t\t}\n")
		} else {
			ctx.writeFieldValueMarshal(sb, "repr", fnum, finfo.ReprType, fopts, field.WriteEmpty, "\t\t")
		}
		sb.WriteString("\t}\n")
		return nil
	}
	if finfo.IsAminoMarshaler && finfo.Type.Kind() == reflect.Struct {
		fmt.Fprintf(sb, "\t{\n")
		fmt.Fprintf(sb, "\t\trepr, err := %s.MarshalAmino()\n", accessor)
		sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
		rinfo := finfo.ReprType
		if rinfo.Type.Kind() == reflect.Struct {
			sb.WriteString("\t\tbefore := offset\n")
			sb.WriteString("\t\toffset, err = repr.MarshalBinary2(cdc, buf, offset)\n")
			sb.WriteString("\t\tif err != nil {\n\t\t\treturn offset, err\n\t\t}\n")
			ctx.writeLengthPrefixedField(sb, fnum, finfo.GetTyp3(fopts), field.WriteEmpty, "\t\t")
		} else {
			// Same repr-zeroness rule as the Go-type-non-struct branch
			// above: only emit when the repr value is non-zero. Without
			// this, a zero std.Coin (whose MarshalAmino returns "")
			// would encode a stray `<fnum> 0x00` that reflect rolls back
			// via writeFieldIfNotEmpty.
			reprZeroCheck := ctx.zeroCheck("repr", rinfo, fopts)
			if reprZeroCheck != "" && !field.WriteEmpty {
				fmt.Fprintf(sb, "\t\tif %s {\n", reprZeroCheck)
				ctx.writeFieldValueMarshal(sb, "repr", fnum, rinfo, fopts, false, "\t\t\t")
				sb.WriteString("\t\t}\n")
			} else {
				ctx.writeFieldValueMarshal(sb, "repr", fnum, rinfo, fopts, field.WriteEmpty, "\t\t")
			}
		}
		sb.WriteString("\t}\n")
		return nil
	}

	// Handle non-pointer fields: skip if default/zero value (unless WriteEmpty).
	if !field.WriteEmpty {
		zeroCheck := ctx.zeroCheck(accessor, finfo, fopts)
		if zeroCheck != "" {
			fmt.Fprintf(sb, "\tif %s {\n", zeroCheck)
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
		fmt.Fprintf(sb, "%sif %s {\n", indent, zeroCheck)
		ctx.writeFieldValueMarshal(sb, accessor, fnum, finfo, fopts, false, indent+"\t")
		fmt.Fprintf(sb, "%s}\n", indent)
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
		fmt.Fprintf(sb, "%s{\n", indent)
		fmt.Fprintf(sb, "%s\tbefore := offset\n", indent)
		fmt.Fprintf(sb, "%s\toffset, err = amino.PrependTime(buf, offset, %s)\n", indent, accessor)
		fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent)
		ctx.writeLengthPrefixedField(sb, fnum, typ3, writeEmpty, indent+"\t")
		fmt.Fprintf(sb, "%s}\n", indent)

	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		fmt.Fprintf(sb, "%s{\n", indent)
		fmt.Fprintf(sb, "%s\tbefore := offset\n", indent)
		fmt.Fprintf(sb, "%s\toffset, err = amino.PrependDuration(buf, offset, %s)\n", indent, accessor)
		fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent)
		ctx.writeLengthPrefixedField(sb, fnum, typ3, writeEmpty, indent+"\t")
		fmt.Fprintf(sb, "%s}\n", indent)

	case rinfo.Type.Kind() == reflect.Struct:
		fmt.Fprintf(sb, "%s{\n", indent)
		fmt.Fprintf(sb, "%s\tbefore := offset\n", indent)
		fmt.Fprintf(sb, "%s\toffset, err = %s.MarshalBinary2(cdc, buf, offset)\n", indent, accessor)
		fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent)
		ctx.writeLengthPrefixedField(sb, fnum, typ3, writeEmpty, indent+"\t")
		fmt.Fprintf(sb, "%s}\n", indent)

	case rinfo.Type.Kind() == reflect.Interface:
		ctx.writeInterfaceFieldMarshal(sb, accessor, fnum, indent)

	case rinfo.Type.Kind() == reflect.String:
		ctx.writeByteLengthFieldWithRollback(sb, fnum, writeEmpty, indent,
			fmt.Sprintf("offset = amino.PrependString(buf, offset, string(%s))", accessor))

	case rinfo.Type.Kind() == reflect.Slice && rinfo.Type.Elem().Kind() == reflect.Uint8:
		ctx.writeByteLengthFieldWithRollback(sb, fnum, writeEmpty, indent,
			fmt.Sprintf("offset = amino.PrependByteSlice(buf, offset, %s)", accessor))

	case rinfo.Type.Kind() == reflect.Array && rinfo.Type.Elem().Kind() == reflect.Uint8:
		ctx.writeByteLengthFieldWithRollback(sb, fnum, writeEmpty, indent,
			fmt.Sprintf("offset = amino.PrependByteSlice(buf, offset, %s[:])", accessor))

	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Packed list (non-byte elements).
		fmt.Fprintf(sb, "%s{\n", indent)
		fmt.Fprintf(sb, "%s\tbefore := offset\n", indent)
		einfo := finfo.Elem
		ert := rinfo.Type.Elem()
		eFopts := fopts
		if einfo != nil && einfo.ReprType.Type.Kind() == reflect.Uint8 {
			fmt.Fprintf( // List of (repr) bytes.
				sb, "%s\tfor i := len(%s) - 1; i >= 0; i-- {\n", indent, accessor)
			fmt.Fprintf(sb, "%s\t\te := %s[i]\n", indent, accessor)
			eAccessor := "e"
			if ert.Kind() == reflect.Ptr {
				fmt.Fprintf(sb, "%s\t\tvar de %s\n", indent, ctx.goTypeName(ert.Elem()))
				fmt.Fprintf(sb, "%s\t\tif e != nil {\n%s\t\t\tde = *e\n%s\t\t}\n", indent, indent, indent)
				eAccessor = "de"
			}
			if einfo.IsAminoMarshaler {
				fmt.Fprintf(sb, "%s\t\ter, err := %s.MarshalAmino()\n", indent, eAccessor)
				fmt.Fprintf(sb, "%s\t\tif err != nil {\n%s\t\t\treturn offset, err\n%s\t\t}\n", indent, indent, indent)
				fmt.Fprintf(sb, "%s\t\toffset = amino.PrependByte(buf, offset, uint8(er))\n", indent)
			} else {
				fmt.Fprintf(sb, "%s\t\toffset = amino.PrependByte(buf, offset, uint8(%s))\n", indent, eAccessor)
			}
			fmt.Fprintf(sb, "%s\t}\n", indent)
		} else if einfo != nil {
			fmt.Fprintf(sb, "%s\tfor i := len(%s) - 1; i >= 0; i-- {\n", indent, accessor)
			fmt.Fprintf(sb, "%s\t\te := %s[i]\n", indent, accessor)
			if ert.Kind() == reflect.Ptr {
				fmt.Fprintf(sb, "%s\t\tif e == nil {\n%s\t\t\te = new(%s)\n%s\t\t}\n", indent, indent, ctx.goTypeName(ert.Elem()), indent)
				if einfo.IsAminoMarshaler {
					fmt.Fprintf(sb, "%s\t\ter, err := (*e).MarshalAmino()\n", indent)
					fmt.Fprintf(sb, "%s\t\tif err != nil {\n%s\t\t\treturn offset, err\n%s\t\t}\n", indent, indent, indent)
					ctx.writePrimitiveEncode(sb, "er", einfo.ReprType, eFopts, indent+"\t\t")
				} else {
					ctx.writePrimitiveEncode(sb, "(*e)", einfo, eFopts, indent+"\t\t")
				}
			} else {
				if einfo.IsAminoMarshaler {
					fmt.Fprintf(sb, "%s\t\ter, err := e.MarshalAmino()\n", indent)
					fmt.Fprintf(sb, "%s\t\tif err != nil {\n%s\t\t\treturn offset, err\n%s\t\t}\n", indent, indent, indent)
					ctx.writePrimitiveEncode(sb, "er", einfo.ReprType, eFopts, indent+"\t\t")
				} else {
					ctx.writePrimitiveEncode(sb, "e", einfo, eFopts, indent+"\t\t")
				}
			}
			fmt.Fprintf(sb, "%s\t}\n", indent)
		}
		// For packed lists, always write (outer len check ensures non-empty).
		ctx.writeLengthPrefixedField(sb, fnum, typ3, true, indent+"\t")
		fmt.Fprintf(sb, "%s}\n", indent)

	default:
		// Primitive types (Bool/Int/Uint/Float/etc., non-length-prefixed):
		// value first, then field key. Mirror reflect writeFieldIfNotEmpty
		// (binary_encode.go:592): when writeEmpty is false, roll back if
		// the emitted value bytes total exactly 1 byte = 0x00. The check
		// operates at the same primitive-value layer as reflect's :592,
		// which is correct for non-ByteLength typ3 fields. Every current
		// caller wraps this branch in an outer zeroCheck so the inner
		// rollback never fires today; defensive hardening only.
		// Sibling of writeReprMarshal's primitive default branch (same
		// layer) and writeByteLengthFieldWithRollback (operates on
		// length-prefixed values).
		if writeEmpty {
			ctx.writePrimitiveEncode(sb, accessor, rinfo, fopts, indent)
			fmt.Fprintf(sb, "%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, %s)\n",
				indent, fnum, typ3GoStr(typ3))
		} else {
			fmt.Fprintf(sb, "%s{\n", indent)
			fmt.Fprintf(sb, "%s\tbefore := offset\n", indent)
			ctx.writePrimitiveEncode(sb, accessor, rinfo, fopts, indent+"\t")
			fmt.Fprintf(sb, "%s\tvalueLen := before - offset\n", indent)
			fmt.Fprintf(sb, "%s\tif valueLen > 1 || (valueLen == 1 && buf[offset] != 0x00) {\n", indent)
			fmt.Fprintf(sb, "%s\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, %s)\n",
				indent, fnum, typ3GoStr(typ3))
			fmt.Fprintf(sb, "%s\t} else {\n", indent)
			fmt.Fprintf(sb, "%s\t\toffset = before\n", indent)
			fmt.Fprintf(sb, "%s\t}\n", indent)
			fmt.Fprintf(sb, "%s}\n", indent)
		}
	}
}

// writeByteLengthFieldWithRollback emits a length-self-prefixing value
// (PrependString / PrependByteSlice) followed by the field key, with a
// post-emission rollback that mirrors reflect's writeFieldIfNotEmpty
// (binary_encode.go:592) when writeEmpty=false. Operates at the
// OUTER-VALUE layer: `valueLen` is the bytes written by `valueStmt`,
// which already includes the length prefix (PrependString writes
// `<uvarint-len><content>`). For an empty string / nil byte slice /
// `[0]byte`, PrependString/ByteSlice emits exactly `[0x00]` (the
// length-0 byte) and the gate fires — same byte position reflect's :592
// would inspect. Sibling of the writeFieldValueMarshal default branch
// rollback (#16), which operates at the same value-bytes layer.
//
// LAYER NOTE: this is intentionally NOT used at writeLengthPrefixedField
// (struct/Time/Duration) or writeInterfaceFieldMarshal (interface)
// sites, because those measure inner-body length BEFORE the prefix is
// added. At those sites, the correct gate is `dataLen > 0` (matches
// reflect's empty-contents case in writeMaybeBare) and a 1-byte body of
// [0x00] correctly does NOT roll back.
func (ctx *P3Context2) writeByteLengthFieldWithRollback(sb *strings.Builder, fnum uint32, writeEmpty bool, indent, valueStmt string) {
	if writeEmpty {
		fmt.Fprintf(sb, "%s%s\n", indent, valueStmt)
		fmt.Fprintf(sb, "%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", indent, fnum)
		return
	}
	fmt.Fprintf(sb, "%s{\n", indent)
	fmt.Fprintf(sb, "%s\tbefore := offset\n", indent)
	fmt.Fprintf(sb, "%s\t%s\n", indent, valueStmt)
	fmt.Fprintf(sb, "%s\tvalueLen := before - offset\n", indent)
	fmt.Fprintf(sb, "%s\tif valueLen > 1 || (valueLen == 1 && buf[offset] != 0x00) {\n", indent)
	fmt.Fprintf(sb, "%s\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", indent, fnum)
	fmt.Fprintf(sb, "%s\t} else {\n", indent)
	fmt.Fprintf(sb, "%s\t\toffset = before\n", indent)
	fmt.Fprintf(sb, "%s\t}\n", indent)
	fmt.Fprintf(sb, "%s}\n", indent)
}

// writeLengthPrefixedField writes the length prefix + field key after data has been
// written backward. Assumes `before` and `offset` are in scope.
//
// LAYER NOTE (re BINARY_FIXES #26): the gate operates at the INNER-BODY layer
// (`dataLen` = bytes of the body BEFORE the length prefix is added). Reflect's
// `writeFieldIfNotEmpty` rollback (binary_encode.go:592) operates at the
// OUTER-VALUE layer (after `writeMaybeBare` has emitted `<uvarint-len><body>`).
// For Typ3ByteLength fields, reflect's :592 rollback fires only when the
// outer-value bytes total exactly 1 byte = 0x00, which corresponds to
// writeMaybeBare's empty-contents branch (writes `[0x00]` as the lone length
// byte) — i.e., dataLen == 0 here. Therefore the correct gate at this layer
// is `dataLen > 0`. A hypothetical 1-byte body of `[0x00]` would be wrapped
// by reflect as `<key> 0x01 0x00` (3 bytes), NOT rolled back. Mirror that.
func (ctx *P3Context2) writeLengthPrefixedField(sb *strings.Builder, fnum uint32, typ3 amino.Typ3, writeEmpty bool, indent string) {
	fmt.Fprintf(sb, "%sdataLen := before - offset\n", indent)
	if !writeEmpty {
		fmt.Fprintf(sb, "%sif dataLen > 0 {\n", indent)
		fmt.Fprintf(sb, "%s\toffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n", indent)
		fmt.Fprintf(sb, "%s\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, %s)\n", indent, fnum, typ3GoStr(typ3))
		fmt.Fprintf(sb, "%s} else {\n", indent)
		fmt.Fprintf(sb, "%s\toffset = before\n", indent)
		fmt.Fprintf(sb, "%s}\n", indent)
	} else {
		fmt.Fprintf(sb, "%soffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n", indent)
		fmt.Fprintf(sb, "%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, %s)\n", indent, fnum, typ3GoStr(typ3))
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
		// Post codec.go UnpackedList fix: beOptionByte is unreachable here
		// (UnpackedList=true requires the element's ReprType.Type kind to
		// map to Typ3ByteLength, excluding Uint8). Assert for future-proofing.
		if beOptionByte {
			return fmt.Errorf("unreachable: writeUnpackedListMarshal reached with beOptionByte=true (type=%v) — codec.go UnpackedList determination should have excluded this", finfo.Type)
		}
		fmt.Fprintf(sb, "\tif len(%s) > 0 {\n", accessor)
		sb.WriteString("\t\tbefore := offset\n")
		fmt.Fprintf(sb, "\t\tfor i := len(%s) - 1; i >= 0; i-- {\n", accessor)
		fmt.Fprintf(sb, "\t\t\telem := %s[i]\n", accessor)
		if ert.Kind() == reflect.Ptr {
			sb.WriteString("\t\t\tif elem == nil {\n")
			sb.WriteString("\t\t\t\telem = new(" + ctx.goTypeName(ert.Elem()) + ")\n")
			sb.WriteString("\t\t\t}\n")
			ctx.writePrimitiveEncode(sb, "(*elem)", einfo, fopts, "\t\t\t")
		} else {
			ctx.writePrimitiveEncode(sb, "elem", einfo, fopts, "\t\t\t")
		}
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tdataLen := before - offset\n")
		sb.WriteString("\t\toffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n")
		fmt.Fprintf(sb, "\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", fopts.BinFieldNum)
		sb.WriteString("\t}\n")
	} else {
		// Unpacked form: repeated field key per element.
		// writeImplicit: nested list whose inner elements are length-prefixed.
		// TypeInfo invariant: einfo.Elem is non-nil for list types; reflect
		// (binary_encode.go:400) dereferences without a nil guard. Match that.
		ertIsPointer := ert.Kind() == reflect.Ptr
		writeImplicit := isListType(einfo.Type) &&
			einfo.Elem.ReprType.Type.Kind() != reflect.Uint8 &&
			einfo.Elem.ReprType.GetTyp3(fopts) != amino.Typ3ByteLength
		fmt.Fprintf(sb, "\tfor i := len(%s) - 1; i >= 0; i-- {\n", accessor)
		fmt.Fprintf(sb, "\t\telem := %s[i]\n", accessor)

		elemAccessor := "elem"
		extraIndent := "\t\t"

		if ertIsPointer {
			// Key off the Go element type (einfo.Type), not the repr type.
			// nil_elements is a Go-side semantic guard: nil `*Struct` is
			// disallowed in lists unless opted in, regardless of the repr's
			// wire kind (per binary_encode.go:399 ertIsStruct check).
			ertIsStruct := einfo.Type.Kind() == reflect.Struct
			sb.WriteString("\t\tif elem == nil {\n")
			if ertIsStruct && !fopts.NilElements {
				sb.WriteString("\t\t\treturn offset, errors.New(\"nil struct pointers in lists not supported unless nil_elements field tag is also set\")\n")
			} else {
				sb.WriteString("\t\t\toffset = amino.PrependByte(buf, offset, 0x00)\n")
				fmt.Fprintf(sb, "\t\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", fopts.BinFieldNum)
			}
			// Mirror reflect's isNonstructDefaultValue pointer recursion
			// (reflect.go:80-86): for non-struct pointer elements, a non-nil
			// pointer to a default value (e.g. *string -> "") also takes
			// the sentinel branch. Wire bytes match today via downstream
			// PrependString("")/PrependByteSlice(nil) producing 0x00, but
			// the explicit branch removes the structural dependency. Skip
			// for AminoMarshaler types (zeroCheck would target the wrong
			// type; reflect's check on the Go-side Kind returns false).
			derefZc := ctx.zeroCheck("(*elem)", einfo, fopts)
			if !ertIsStruct && !einfo.IsAminoMarshaler && derefZc != "" {
				fmt.Fprintf(sb, "\t\t} else if !(%s) {\n", derefZc)
				sb.WriteString("\t\t\toffset = amino.PrependByte(buf, offset, 0x00)\n")
				fmt.Fprintf(sb, "\t\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", fopts.BinFieldNum)
			}
			sb.WriteString("\t\t} else {\n")
			elemAccessor = "(*elem)"
			extraIndent = "\t\t\t"
		}

		if einfo.ReprType.Type.Kind() == reflect.Interface {
			fmt.Fprintf( // Interface element: encode via MarshalAnyBinary2.
				sb, "%sif %s != nil {\n", extraIndent, elemAccessor)
			fmt.Fprintf(sb, "%s\tbefore := offset\n", extraIndent)
			fmt.Fprintf(sb, "%s\toffset, err = cdc.MarshalAnyBinary2(%s, buf, offset)\n", extraIndent, elemAccessor)
			fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", extraIndent, extraIndent, extraIndent)
			fmt.Fprintf(sb, "%s\tanyLen := before - offset\n", extraIndent)
			fmt.Fprintf(sb, "%s\toffset = amino.PrependUvarint(buf, offset, uint64(anyLen))\n", extraIndent)
			fmt.Fprintf(sb, "%s} else {\n", extraIndent)
			fmt.Fprintf(sb, "%s\toffset = amino.PrependByte(buf, offset, 0x00)\n", extraIndent)
			fmt.Fprintf(sb, "%s}\n", extraIndent)
			fmt.Fprintf(sb, "%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", extraIndent, fopts.BinFieldNum)
		} else if writeImplicit {
			fmt.Fprintf( // Nested list: wrap in implicit struct.
				sb, "%s{\n", extraIndent)
			fmt.Fprintf(sb, "%s\timplicitBefore := offset\n", extraIndent)
			// Encode inner list elements.
			ctx.writeListEncode(sb, elemAccessor, einfo, fopts, extraIndent+"\t")
			fmt.Fprintf(sb, "%s\tinnerLen := implicitBefore - offset\n", extraIndent)
			fmt.Fprintf(sb, "%s\tif innerLen > 0 {\n", extraIndent)
			fmt.Fprintf(sb, "%s\t\toffset = amino.PrependUvarint(buf, offset, uint64(innerLen))\n", extraIndent)
			fmt.Fprintf(sb, "%s\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n", extraIndent)
			fmt.Fprintf(sb, "%s\t}\n", extraIndent)
			fmt.Fprintf( // Outer: compute implicit struct size and write length prefix + field key.
				sb, "%s\tissLen := implicitBefore - offset\n", extraIndent)
			fmt.Fprintf(sb, "%s\toffset = amino.PrependUvarint(buf, offset, uint64(issLen))\n", extraIndent)
			fmt.Fprintf(sb, "%s}\n", extraIndent)
			fmt.Fprintf(sb, "%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", extraIndent, fopts.BinFieldNum)
		} else if einfo.ReprType.Type.Kind() == reflect.Struct ||
			einfo.ReprType.Type == reflect.TypeOf(time.Duration(0)) ||
			(isListType(einfo.ReprType.Type) && einfo.ReprType.Type.Elem().Kind() != reflect.Uint8) {
			fmt.Fprintf( // Struct/Duration/nested-list element: encode backward, then length-prefix.
				sb, "%sbefore := offset\n", extraIndent)
			ctx.writeElementEncode(sb, elemAccessor, einfo, fopts, extraIndent)
			fmt.Fprintf(sb, "%sdataLen := before - offset\n", extraIndent)
			fmt.Fprintf(sb, "%soffset = amino.PrependUvarint(buf, offset, uint64(dataLen))\n", extraIndent)
			fmt.Fprintf(sb, "%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", extraIndent, fopts.BinFieldNum)
		} else {
			// Non-struct ByteLength element (string, []byte). Mirror reflect's
			// isNonstructDefaultValue branch (binary_encode.go:416-432): when the
			// non-pointer element is default (e.g. "" string, nil/empty []byte),
			// reflect emits a direct 0x00 sentinel rather than length-prefixed
			// content. Wire bytes are identical (PrependString("") also emits a
			// single 0x00 length-byte), but the explicit branch removes the
			// dependency on downstream length-emission composition — a future
			// refactor of writeElementEncode that changed sentinel formation
			// would otherwise silently regress.
			//
			// Skip for AminoMarshaler elements: zeroCheck uses ReprType, so the
			// emitted comparison would be against the wrong Go type. Reflect's
			// isNonstructDefaultValue keys off the GO-SIDE Kind (Array/Struct
			// for AminoMarshaler types) and returns false anyway, so no
			// sentinel branch is taken there in either codec.
			zc := ctx.zeroCheck(elemAccessor, einfo, fopts)
			if zc != "" && !ertIsPointer && !einfo.IsAminoMarshaler {
				fmt.Fprintf(sb, "%sif %s {\n", extraIndent, zc)
				ctx.writeElementEncode(sb, elemAccessor, einfo, fopts, extraIndent+"\t")
				fmt.Fprintf(sb, "%s} else {\n", extraIndent)
				fmt.Fprintf(sb, "%s\toffset = amino.PrependByte(buf, offset, 0x00)\n", extraIndent)
				fmt.Fprintf(sb, "%s}\n", extraIndent)
			} else {
				ctx.writeElementEncode(sb, elemAccessor, einfo, fopts, extraIndent)
			}
			fmt.Fprintf(sb, "%soffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", extraIndent, fopts.BinFieldNum)
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
		fmt.Fprintf(sb, "%sfor i := len(%s) - 1; i >= 0; i-- {\n", indent, accessor)
		fmt.Fprintf(sb, "%s\te := %s[i]\n", indent, accessor)
		if ert.Kind() == reflect.Ptr {
			fmt.Fprintf(sb, "%s\tif e == nil {\n%s\t\te = new(%s)\n%s\t}\n", indent, indent, ctx.goTypeName(ert.Elem()), indent)
			ctx.writePrimitiveEncode(sb, "(*e)", einfo, fopts, indent+"\t")
		} else {
			ctx.writePrimitiveEncode(sb, "e", einfo, fopts, indent+"\t")
		}
		fmt.Fprintf(sb, "%s}\n", indent)
	} else {
		fmt.Fprintf(sb, "%sfor i := len(%s) - 1; i >= 0; i-- {\n", indent, accessor)
		fmt.Fprintf(sb, "%s\te := %s[i]\n", indent, accessor)
		fmt.Fprintf(sb, "%s\telbefore := offset\n", indent)
		ctx.writeElementEncode(sb, "e", einfo, fopts, indent+"\t")
		fmt.Fprintf(sb, "%s\telLen := elbefore - offset\n", indent)
		fmt.Fprintf(sb, "%s\toffset = amino.PrependUvarint(buf, offset, uint64(elLen))\n", indent)
		fmt.Fprintf(sb, "%s\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n", indent)
		fmt.Fprintf(sb, "%s}\n", indent)
	}
}

func (ctx *P3Context2) writeElementEncode(sb *strings.Builder, accessor string, einfo *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	rinfo := einfo.ReprType
	switch {
	case rinfo.Type == reflect.TypeOf(time.Time{}):
		fmt.Fprintf(sb, "%soffset, err = amino.PrependTime(buf, offset, %s)\n", indent, accessor)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn offset, err\n%s}\n", indent, indent, indent)
	case rinfo.Type == reflect.TypeOf(time.Duration(0)):
		fmt.Fprintf(sb, "%soffset, err = amino.PrependDuration(buf, offset, %s)\n", indent, accessor)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn offset, err\n%s}\n", indent, indent, indent)
	case rinfo.Type.Kind() == reflect.Struct:
		fmt.Fprintf(sb, "%soffset, err = %s.MarshalBinary2(cdc, buf, offset)\n", indent, accessor)
		fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn offset, err\n%s}\n", indent, indent, indent)
	case isListType(rinfo.Type) && rinfo.Type.Elem().Kind() != reflect.Uint8:
		// Nested list element: encode as implicit struct content inline.
		// The caller handles the outer length prefix and field key.
		innerEinfo := einfo.Elem
		if innerEinfo == nil {
			fmt.Fprintf( // Fallback for unexpected cases.
				sb, "%s{\n", indent)
			fmt.Fprintf(sb, "%s\tbz, merr := cdc.Marshal(%s)\n", indent, accessor)
			fmt.Fprintf(sb, "%s\tif merr != nil {\n%s\t\treturn offset, merr\n%s\t}\n", indent, indent, indent)
			fmt.Fprintf(sb, "%s\toffset = amino.PrependBytes(buf, offset, bz)\n", indent)
			fmt.Fprintf(sb, "%s}\n", indent)
			return
		}
		innerTyp3 := innerEinfo.GetTyp3(fopts)
		innerRinfo := innerEinfo.ReprType
		if innerTyp3 != amino.Typ3ByteLength {
			fmt.Fprintf( // Packed inner elements: single field 1 + length prefix.
				sb, "%s{\n", indent)
			fmt.Fprintf(sb, "%s\tpkBefore := offset\n", indent)
			fmt.Fprintf(sb, "%s\tfor ii := len(%s) - 1; ii >= 0; ii-- {\n", indent, accessor)
			fmt.Fprintf(sb, "%s\t\tie := %s[ii]\n", indent, accessor)
			ctx.writePrimitiveEncode(sb, "ie", innerEinfo, fopts, indent+"\t\t")
			fmt.Fprintf(sb, "%s\t}\n", indent)
			fmt.Fprintf(sb, "%s\tpkLen := pkBefore - offset\n", indent)
			fmt.Fprintf(sb, "%s\tif pkLen > 0 {\n", indent)
			fmt.Fprintf(sb, "%s\t\toffset = amino.PrependUvarint(buf, offset, uint64(pkLen))\n", indent)
			fmt.Fprintf(sb, "%s\t\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n", indent)
			fmt.Fprintf(sb, "%s\t}\n", indent)
			fmt.Fprintf(sb, "%s}\n", indent)
		} else {
			fmt.Fprintf( // ByteLength inner elements: repeated field 1 entries.
				sb, "%sfor ii := len(%s) - 1; ii >= 0; ii-- {\n", indent, accessor)
			fmt.Fprintf(sb, "%s\tie := %s[ii]\n", indent, accessor)
			switch {
			case innerRinfo.Type == reflect.TypeOf(time.Time{}):
				fmt.Fprintf(sb, "%s\tieBefore := offset\n", indent)
				fmt.Fprintf(sb, "%s\toffset, err = amino.PrependTime(buf, offset, ie)\n", indent)
				fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent)
				fmt.Fprintf(sb, "%s\tieLen := ieBefore - offset\n", indent)
				fmt.Fprintf(sb, "%s\toffset = amino.PrependUvarint(buf, offset, uint64(ieLen))\n", indent)
			case innerRinfo.Type == reflect.TypeOf(time.Duration(0)):
				fmt.Fprintf(sb, "%s\tieBefore := offset\n", indent)
				fmt.Fprintf(sb, "%s\toffset, err = amino.PrependDuration(buf, offset, ie)\n", indent)
				fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent)
				fmt.Fprintf(sb, "%s\tieLen := ieBefore - offset\n", indent)
				fmt.Fprintf(sb, "%s\toffset = amino.PrependUvarint(buf, offset, uint64(ieLen))\n", indent)
			case innerRinfo.Type.Kind() == reflect.Struct:
				fmt.Fprintf(sb, "%s\tieBefore := offset\n", indent)
				fmt.Fprintf(sb, "%s\toffset, err = ie.MarshalBinary2(cdc, buf, offset)\n", indent)
				fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent)
				fmt.Fprintf(sb, "%s\tieLen := ieBefore - offset\n", indent)
				fmt.Fprintf(sb, "%s\toffset = amino.PrependUvarint(buf, offset, uint64(ieLen))\n", indent)
			default:
				// String, []byte: writePrimitiveEncode includes own length prefix.
				ctx.writePrimitiveEncode(sb, "ie", innerEinfo, fopts, indent+"\t")
			}
			fmt.Fprintf(sb, "%s\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, 1, amino.Typ3ByteLength)\n", indent)
			fmt.Fprintf(sb, "%s}\n", indent)
		}
	default:
		if einfo.IsAminoMarshaler {
			fmt.Fprintf(sb, "%ser, err := %s.MarshalAmino()\n", indent, accessor)
			fmt.Fprintf(sb, "%sif err != nil {\n%s\treturn offset, err\n%s}\n", indent, indent, indent)
			ctx.writePrimitiveEncode(sb, "er", einfo.ReprType, fopts, indent)
		} else {
			ctx.writePrimitiveEncode(sb, accessor, einfo, fopts, indent)
		}
	}
}

// writeInterfaceFieldMarshal emits MarshalAnyBinary2 + length prefix + key for
// a non-nil interface field. There is no single-0x00 outer rollback gate here:
// `anyLen` is the size of the inner Any body BEFORE the length prefix is
// added; reflect's writeFieldIfNotEmpty rollback (binary_encode.go:592) fires
// at the OUTER-VALUE layer (post-prefix) and only when the value bytes total
// exactly 1 byte = 0x00, which for Typ3ByteLength happens only when the
// inner body is empty (writeMaybeBare's empty-contents branch). For interface
// fields, MarshalAnyBinary2 always emits at minimum a TypeURL field (≥ 6
// bytes), so anyLen ≥ 6 always; an empty/single-0x00 body case isn't reachable.
// See LAYER NOTE on writeLengthPrefixedField above for the layer-mismatch
// rationale.
func (ctx *P3Context2) writeInterfaceFieldMarshal(sb *strings.Builder, accessor string, fnum uint32, indent string) {
	fmt.Fprintf(sb, "%sif %s != nil {\n", indent, accessor)
	fmt.Fprintf(sb, "%s\tbefore := offset\n", indent)
	fmt.Fprintf(sb, "%s\toffset, err = cdc.MarshalAnyBinary2(%s, buf, offset)\n", indent, accessor)
	fmt.Fprintf(sb, "%s\tif err != nil {\n%s\t\treturn offset, err\n%s\t}\n", indent, indent, indent)
	fmt.Fprintf(sb, "%s\tanyLen := before - offset\n", indent)
	fmt.Fprintf(sb, "%s\toffset = amino.PrependUvarint(buf, offset, uint64(anyLen))\n", indent)
	fmt.Fprintf(sb, "%s\toffset = amino.PrependFieldNumberAndTyp3(buf, offset, %d, amino.Typ3ByteLength)\n", indent, fnum)
	fmt.Fprintf(sb, "%s}\n", indent)
}

// === Primitive Encoding ===

func (ctx *P3Context2) writePrimitiveEncode(sb *strings.Builder, accessor string, info *amino.TypeInfo, fopts amino.FieldOptions, indent string) {
	rinfo := info.ReprType
	rt := rinfo.Type
	kind := rt.Kind()

	switch kind {
	case reflect.Bool:
		fmt.Fprintf(sb, "%soffset = amino.PrependBool(buf, offset, bool(%s))\n", indent, accessor)
	case reflect.Int8, reflect.Int16:
		fmt.Fprintf(sb, "%soffset = amino.PrependVarint(buf, offset, int64(%s))\n", indent, accessor)
	case reflect.Int32:
		if fopts.BinFixed32 {
			fmt.Fprintf(sb, "%soffset = amino.PrependInt32(buf, offset, int32(%s))\n", indent, accessor)
		} else {
			fmt.Fprintf(sb, "%soffset = amino.PrependVarint(buf, offset, int64(%s))\n", indent, accessor)
		}
	case reflect.Int64:
		if fopts.BinFixed64 {
			fmt.Fprintf(sb, "%soffset = amino.PrependInt64(buf, offset, int64(%s))\n", indent, accessor)
		} else {
			fmt.Fprintf(sb, "%soffset = amino.PrependVarint(buf, offset, int64(%s))\n", indent, accessor)
		}
	case reflect.Int:
		if fopts.BinFixed64 {
			fmt.Fprintf(sb, "%soffset = amino.PrependInt64(buf, offset, int64(%s))\n", indent, accessor)
		} else if fopts.BinFixed32 {
			fmt.Fprintf(sb, "%soffset = amino.PrependInt32(buf, offset, int32(%s))\n", indent, accessor)
		} else {
			fmt.Fprintf(sb, "%soffset = amino.PrependVarint(buf, offset, int64(%s))\n", indent, accessor)
		}
	case reflect.Uint8:
		fmt.Fprintf(sb, "%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor)
	case reflect.Uint16:
		fmt.Fprintf(sb, "%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor)
	case reflect.Uint32:
		if fopts.BinFixed32 {
			fmt.Fprintf(sb, "%soffset = amino.PrependUint32(buf, offset, uint32(%s))\n", indent, accessor)
		} else {
			fmt.Fprintf(sb, "%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor)
		}
	case reflect.Uint64:
		if fopts.BinFixed64 {
			fmt.Fprintf(sb, "%soffset = amino.PrependUint64(buf, offset, uint64(%s))\n", indent, accessor)
		} else {
			fmt.Fprintf(sb, "%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor)
		}
	case reflect.Uint:
		if fopts.BinFixed64 {
			fmt.Fprintf(sb, "%soffset = amino.PrependUint64(buf, offset, uint64(%s))\n", indent, accessor)
		} else if fopts.BinFixed32 {
			fmt.Fprintf(sb, "%soffset = amino.PrependUint32(buf, offset, uint32(%s))\n", indent, accessor)
		} else {
			fmt.Fprintf(sb, "%soffset = amino.PrependUvarint(buf, offset, uint64(%s))\n", indent, accessor)
		}
	case reflect.Float32:
		// Mirror reflect (binary_encode.go:193-197): floats require
		// amino:"unsafe". ValidateBasic (codec.go:166-175) panics at
		// codec-init for registered types, so this is unreachable via
		// the registry — but emission-site enforcement keeps the
		// generator's contract explicit.
		if !fopts.Unsafe {
			panic(fmt.Sprintf("genproto2: writePrimitiveEncode: float32 emission requires amino:\"unsafe\" (type=%v, accessor=%s)", rt, accessor))
		}
		fmt.Fprintf(sb, "%soffset = amino.PrependFloat32(buf, offset, float32(%s))\n", indent, accessor)
	case reflect.Float64:
		if !fopts.Unsafe {
			panic(fmt.Sprintf("genproto2: writePrimitiveEncode: float64 emission requires amino:\"unsafe\" (type=%v, accessor=%s)", rt, accessor))
		}
		fmt.Fprintf(sb, "%soffset = amino.PrependFloat64(buf, offset, float64(%s))\n", indent, accessor)
	case reflect.String:
		fmt.Fprintf(sb, "%soffset = amino.PrependString(buf, offset, string(%s))\n", indent, accessor)
	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			fmt.Fprintf(sb, "%soffset = amino.PrependByteSlice(buf, offset, %s)\n", indent, accessor)
		} else {
			panic(fmt.Sprintf("genproto2: writePrimitiveEncode: unsupported slice element kind %v (type=%v, accessor=%s)", rt.Elem().Kind(), rt, accessor))
		}
	case reflect.Array:
		if rt.Elem().Kind() == reflect.Uint8 {
			fmt.Fprintf(sb, "%soffset = amino.PrependByteSlice(buf, offset, %s[:])\n", indent, accessor)
		} else {
			panic(fmt.Sprintf("genproto2: writePrimitiveEncode: unsupported array element kind %v (type=%v, accessor=%s)", rt.Elem().Kind(), rt, accessor))
		}
	default:
		panic(fmt.Sprintf("genproto2: writePrimitiveEncode: unsupported kind %v (type=%v, accessor=%s)", kind, rt, accessor))
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
		return accessor // bool is truthy check
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%s != 0", accessor)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%s != 0", accessor)
	case reflect.Float32, reflect.Float64:
		// Float is never "default" per isNonstructDefaultValue (reflect.go:101
		// falls to the default false branch). Reflect emits zero Float as 4
		// or 8 zero bytes; the generator must match. Return "" → no zero-skip.
		return ""
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

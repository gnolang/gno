package amino

import (
	"fmt"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

const bdOptionByte = 0x01

// ----------------------------------------
// cdc.decodeReflectBinary

// This is the main entrypoint for decoding all types from binary form. This
// function calls decodeReflectBinary*, and generally those functions should
// only call this one, for overrides all happen here.
//
// "bare" is ignored when the value is a primitive type, or a byteslice or
// bytearray, but this is confusing and should probably be improved with
// explicit expectations.
//
// This function will always construct an instance if rv.Kind() is pointer,
// even if there is nothing left to read.  If a nil value is desired,
// decodeReflectBinary() should not be called on rv.
//
// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinary(bz []byte, info *TypeInfo,
	rv reflect.Value, fopts FieldOptions, bare bool, options uint64,
) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if info.Type.Kind() == reflect.Interface && rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}
	if printLog {
		fmt.Printf("(D) decodeReflectBinary(bz: %X, info: %v, rv: %#v (%v), fopts: %v)\n",
			bz, info, rv.Interface(), rv.Type(), fopts)
		defer func() {
			fmt.Printf("(D) -> n: %v, err: %v\n", n, err)
		}()
	}
	var _n int

	// Dereference-and-construct if pointer.
	rv = maybeDerefAndConstruct(rv)

	// Handle the most special case, "well known".
	if info.IsBinaryWellKnownType {
		var ok bool
		ok, n, err = decodeReflectBinaryWellKnown(bz, info, rv, fopts, bare)
		if ok || err != nil {
			return
		}
	}

	// Handle override if a pointer to rv implements UnmarshalAmino.
	if info.IsAminoMarshaler {
		// First, decode repr instance from bytes.
		rrv := reflect.New(info.ReprType.Type).Elem()
		var rinfo *TypeInfo
		rinfo, err = cdc.getTypeInfoWLock(info.ReprType.Type)
		if err != nil {
			return
		}
		_n, err = cdc.decodeReflectBinary(bz, rinfo, rrv, fopts, bare, options)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		// Then, decode from repr instance.
		uwrm := rv.Addr().MethodByName("UnmarshalAmino")
		uwouts := uwrm.Call([]reflect.Value{rrv})
		erri := uwouts[0].Interface()
		if erri != nil {
			err = erri.(error)
		}
		return
	}

	switch info.Type.Kind() {
	// ----------------------------------------
	// Complex

	case reflect.Interface:
		_n, err = cdc.decodeReflectBinaryInterface(bz, info, rv, fopts, bare)
		n += _n
		return

	case reflect.Array:
		ert := info.Type.Elem()
		if ert.Kind() == reflect.Uint8 {
			_n, err = cdc.decodeReflectBinaryByteArray(bz, info, rv, fopts)
			n += _n
		} else {
			_n, err = cdc.decodeReflectBinaryArray(bz, info, rv, fopts, bare)
			n += _n
		}
		return

	case reflect.Slice:
		ert := info.Type.Elem()
		if ert.Kind() == reflect.Uint8 {
			_n, err = cdc.decodeReflectBinaryByteSlice(bz, info, rv, fopts)
			n += _n
		} else {
			_n, err = cdc.decodeReflectBinarySlice(bz, info, rv, fopts, bare)
			n += _n
		}
		return

	case reflect.Struct:
		_n, err = cdc.decodeReflectBinaryStruct(bz, info, rv, fopts, bare)
		n += _n
		return

	// ----------------------------------------
	// Signed

	case reflect.Int64:
		var num int64
		if fopts.BinFixed64 {
			num, _n, err = DecodeInt64(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetInt(num)
		} else {
			var u64 int64
			u64, _n, err = DecodeVarint(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetInt(u64)
		}
		return

	case reflect.Int32:
		if fopts.BinFixed32 {
			var num int32
			num, _n, err = DecodeInt32(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetInt(int64(num))
		} else {
			var num int64
			num, _n, err = DecodeVarint(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetInt(num)
		}
		return

	case reflect.Int16:
		var num int16
		num, _n, err = DecodeVarint16(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int8:
		var num int8
		num, _n, err = DecodeVarint8(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(int64(num))
		return

	case reflect.Int:
		var num int64
		num, _n, err = DecodeVarint(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetInt(num)
		return

	// ----------------------------------------
	// Unsigned

	case reflect.Uint64:
		var num uint64
		if fopts.BinFixed64 {
			num, _n, err = DecodeUint64(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetUint(num)
		} else {
			num, _n, err = DecodeUvarint(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetUint(num)
		}
		return

	case reflect.Uint32:
		if fopts.BinFixed32 {
			var num uint32
			num, _n, err = DecodeUint32(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetUint(uint64(num))
		} else {
			var num uint64
			num, _n, err = DecodeUvarint(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			rv.SetUint(num)
		}
		return

	case reflect.Uint16:
		var num uint16
		num, _n, err = DecodeUvarint16(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint8:
		var num uint8
		if options&bdOptionByte != 0 {
			num, _n, err = DecodeByte(bz)
		} else {
			num, _n, err = DecodeUvarint8(bz)
		}
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(uint64(num))
		return

	case reflect.Uint:
		var num uint64
		num, _n, err = DecodeUvarint(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetUint(num)
		return

	// ----------------------------------------
	// Misc.

	case reflect.Bool:
		var b bool
		b, _n, err = DecodeBool(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetBool(b)
		return

	case reflect.Float64:
		var f float64
		if !fopts.Unsafe {
			err = errors.New("float support requires `amino:\"unsafe\"`")
			return
		}
		f, _n, err = DecodeFloat64(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetFloat(f)
		return

	case reflect.Float32:
		var f float32
		if !fopts.Unsafe {
			err = errors.New("float support requires `amino:\"unsafe\"`")
			return
		}
		f, _n, err = DecodeFloat32(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetFloat(float64(f))
		return

	case reflect.String:
		var str string
		str, _n, err = DecodeString(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		rv.SetString(str)
		return

	default:
		panic(fmt.Sprintf("unknown field type %v", info.Type.Kind()))
	}
}

// CONTRACT: rv.CanAddr() is true.
// CONTRACT: rv.Kind() == reflect.Interface.
func (cdc *Codec) decodeReflectBinaryInterface(bz []byte, iinfo *TypeInfo, rv reflect.Value,
	fopts FieldOptions, bare bool,
) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectBinaryInterface")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}
	if !rv.IsNil() {
		// JAE: Heed this note, this is very tricky.
		// I've forgotten the reason a second time,
		// but I'm pretty sure that reason exists.
		err = errors.New("decoding to a non-nil interface is not supported yet")
		return
	}

	// Strip if needed.
	bz, err = decodeMaybeBare(bz, &n, bare)
	if err != nil {
		return
	}

	// Special case if nil interface.
	if len(bz) == 0 {
		rv.Set(iinfo.ZeroValue)
		return
	}

	// Consume first field of TypeURL.
	fnum, typ, _n, err := decodeFieldNumberAndTyp3(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if fnum != 1 || typ != Typ3ByteLength {
		err = fmt.Errorf("expected Any field number 1 TypeURL, got num %v typ %v", fnum, typ)
		return
	}

	// Consume string.
	typeURL, _n, err := DecodeString(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}

	// Consume second field of Value.
	var value []byte = nil
	lenbz := len(bz)
	if lenbz == 0 {
		// Value is empty.
	} else {
		// Consume field key.
		fnum, typ, _n, err = decodeFieldNumberAndTyp3(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		// fnum of greater than 2 is malformed for google.protobuf.Any,
		// and not supported (will error).
		if fnum != 2 || typ != Typ3ByteLength {
			err = fmt.Errorf("expected Any field number 2 Value, got num %v typ %v", fnum, typ)
			return
		}
		// Decode second field value of Value
		// Consume second field value, a byteslice.
		lenbz := len(bz)
		value, _n, err = DecodeByteSlice(bz)
		if slide(&bz, nil, _n) && err != nil {
			return
		}
		// Earlier, we set bz to the byteslice read from
		// buf.  Ensure that all of bz was consumed.
		if len(bz) > 0 {
			err = errors.New("bytes left over after reading Any.")
			return
		}
		// Increment n by length of length prefix for
		// value.  This lets us return the correct *n
		// read.
		n += lenbz - len(value)
	}

	// Decode typeURL and value to rv.
	_n, err = cdc.decodeReflectBinaryAny(typeURL, value, rv, fopts)
	if slide(&value, &n, _n) && err != nil {
		return
	}

	return
}

// Returns the number of bytes read from value.
// CONTRACT: rv.CanAddr() is true.
// CONTRACT: rv.Kind() == reflect.Interface.
func (cdc *Codec) decodeReflectBinaryAny(typeURL string, value []byte, rv reflect.Value, fopts FieldOptions) (n int, err error) {
	// Invalid typeURL value is invalid.
	if !IsASCIIText(typeURL) {
		err = fmt.Errorf("invalid type_url string bytes %X", typeURL)
		return
	}

	// Get concrete type info from typeURL.
	// (we don't consume the value bytes yet).
	var cinfo *TypeInfo
	cinfo, err = cdc.getTypeInfoFromTypeURLRLock(typeURL, fopts)
	if err != nil {
		return
	}

	// Construct the concrete type value.
	crv, irvSet := constructConcreteType(cinfo)

	// Special case when value is default empty value.
	// NOTE: For compatibility with other languages,
	// nil-pointer interface values are forbidden.
	if len(value) == 0 {
		// Verify that the decoded concrete type is assignable to the target interface.
		if !irvSet.Type().AssignableTo(rv.Type()) {
			err = fmt.Errorf("decoded type %v is not assignable to interface %v", irvSet.Type(), rv.Type())
			return
		}
		rv.Set(irvSet)
		return
	}

	// Now fopts field number (for unpacked lists) is reset to 1,
	// otherwise the rest of the field options are inherited.  NOTE:
	// make a function to abstract this.
	fopts.BinFieldNum = 1

	// See if we need to read the typ3 encoding of an implicit struct.
	// google.protobuf.Any values must be a struct, or an unpacked list which
	// is indistinguishable from a struct.
	//
	// See corresponding encoding message in this file, and also
	// Codec.Unmarshal()
	bareValue := true
	if !cinfo.IsStructOrUnpacked(fopts) &&
		len(value) > 0 {
		// TODO test for when !cinfo.IsStructOrUnpacked() but fopts.BinFieldNum != 1.
		var (
			fnum      uint32
			typ       Typ3
			nFnumTyp3 int
		)
		fnum, typ, nFnumTyp3, err = decodeFieldNumberAndTyp3(value)
		if err != nil {
			return n, errors.Wrap(err, "could not decode field number and type")
		}
		if fnum != 1 {
			return n, fmt.Errorf("expected field number: 1; got: %v", fnum)
		}
		typWanted := cinfo.GetTyp3(FieldOptions{})
		if typ != typWanted {
			return n, fmt.Errorf("expected field type %v for # %v of %v, got %v",
				typWanted, fnum, cinfo.Type, typ)
		}
		slide(&value, &n, nFnumTyp3)
		// We have written the implicit struct field key.  Now what
		// follows should be encoded with bare=false, though if typ3 !=
		// Typ3ByteLength, bare is ignored anyways.
		bareValue = false
	}

	// Decode into the concrete type.
	// Here is where we consume the value bytes, which are necessarily length
	// prefixed, due to the type of field 2, so bareValue is false.
	_n, err := cdc.decodeReflectBinary(value, cinfo, crv, fopts, bareValue, 0)
	if slide(&value, &n, _n) && err != nil {
		// Verify that the decoded concrete type is assignable to the target interface.
		// This prevents panics when a registered type doesn't implement the target interface.
		if !irvSet.Type().AssignableTo(rv.Type()) {
			err = fmt.Errorf("decoded type %v is not assignable to interface %v", irvSet.Type(), rv.Type())
			return
		}
		rv.Set(irvSet) // Helps with debugging
		return
	}

	// Ensure that all of value was consumed.
	if len(value) > 0 {
		err = errors.New("bytes left over after reading Any.Value.")
		return
	}

	// We need to set here, for when !PointerPreferred and the type
	// is say, an array of bytes (e.g. [32]byte), then we must call
	// rv.Set() *after* the value was acquired.
	// Verify that the decoded concrete type is assignable to the target interface.
	// This prevents panics when a registered type doesn't implement the target interface.
	if !irvSet.Type().AssignableTo(rv.Type()) {
		err = fmt.Errorf("decoded type %v is not assignable to interface %v", irvSet.Type(), rv.Type())
		return
	}
	// NOTE: rv.Set() should succeed because it was validated
	// already during Register[Interface/Concrete].
	rv.Set(irvSet)
	return n, err
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryByteArray(bz []byte, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions,
) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectBinaryByteArray")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}
	length := info.Type.Len()
	if len(bz) < length {
		return 0, fmt.Errorf("insufficient bytes to decode [%v]byte", length)
	}

	// Read byte-length prefixed byteslice.
	byteslice, _n, err := DecodeByteSlice(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if len(byteslice) != length {
		err = fmt.Errorf("mismatched byte array length: Expected %v, got %v",
			length, len(byteslice))
		return
	}

	// Copy read byteslice to rv array.
	reflect.Copy(rv, reflect.ValueOf(byteslice))
	return n, err
}

// CONTRACT: rv.CanAddr() is true.
// NOTE: Keep the code structure similar to decodeReflectBinarySlice.
func (cdc *Codec) decodeReflectBinaryArray(bz []byte, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions, bare bool,
) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectBinaryArray")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}
	length := info.Type.Len()
	einfo, err := cdc.getTypeInfoWLock(ert)
	if err != nil {
		return
	}

	// Bare if needed.
	bz, err = decodeMaybeBare(bz, &n, bare)
	if err != nil {
		return
	}

	// If elem is not already a ByteLength type, read in packed form.
	// This is a Proto wart due to Proto backwards compatibility issues.
	// Amino2 will probably migrate to use the List typ3.
	newoptions := uint64(0)
	// Special case for list of (repr) bytes: decode from "bytes".
	if ert.Kind() == reflect.Ptr && ert.Elem().Kind() == reflect.Uint8 {
		newoptions |= bdOptionByte
	}
	typ3 := einfo.GetTyp3(fopts)
	if typ3 != Typ3ByteLength || (newoptions&beOptionByte > 0) {
		// Read elements in packed form.
		for i := range length {
			erv := rv.Index(i)
			var _n int
			_n, err = cdc.decodeReflectBinary(bz, einfo, erv, fopts, false, newoptions)
			if slide(&bz, &n, _n) && err != nil {
				err = fmt.Errorf("error reading array contents: %w", err)
				return
			}
		}
		// Ensure that we read the whole buffer.
		if len(bz) > 0 {
			err = errors.New("bytes left over after reading array contents")
			return
		}
	} else {
		// NOTE: ert is for the element value, while einfo.Type is dereferenced.
		isErtStructPointer := ert.Kind() == reflect.Ptr && einfo.Type.Kind() == reflect.Struct
		writeImplicit := isListType(einfo.Type) &&
			einfo.Elem.ReprType.Type.Kind() != reflect.Uint8 &&
			einfo.Elem.ReprType.GetTyp3(fopts) != Typ3ByteLength

		// Read elements in unpacked form.
		for i := range length {
			// Read field key (number and type).
			var (
				fnum uint32
				typ  Typ3
				_n   int
			)
			fnum, typ, _n, err = decodeFieldNumberAndTyp3(bz)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			// Validate field number and typ3.
			if fnum != fopts.BinFieldNum {
				err = errors.New(fmt.Sprintf("expected repeated field number %v, got %v", fopts.BinFieldNum, fnum))
				return
			}
			if typ != Typ3ByteLength {
				err = errors.New(fmt.Sprintf("expected repeated field type %v, got %v", Typ3ByteLength, typ))
				return
			}
			// Decode the next ByteLength bytes into erv.
			erv := rv.Index(i)
			// Special case if:
			//  * next ByteLength bytes are 0x00, and
			//  * - erv is not a struct pointer, or
			//    - field option has NilElements set
			if (len(bz) > 0 && bz[0] == 0x00) &&
				(!isErtStructPointer || fopts.NilElements) {
				slide(&bz, &n, 1)
				erv.Set(defaultValue(erv.Type()))
				continue
			}
			// Special case: nested lists.
			// Multidimensional lists (nested inner lists also in unpacked
			// form) are represented as lists of implicit structs.
			if writeImplicit {
				// Read bytes for implicit struct.
				var ibz []byte
				ibz, _n, err = DecodeByteSlice(bz)
				if slide(&bz, nil, _n) && err != nil {
					return
				}
				// This is a trick for debuggability -- we slide on &n more later.
				n += UvarintSize(uint64(len(ibz)))
				// Read field key of implicit struct.
				var fnum uint32
				fnum, _, _n, err = decodeFieldNumberAndTyp3(ibz)
				if slide(&ibz, &n, _n) && err != nil {
					return
				}
				if fnum != 1 {
					err = fmt.Errorf("unexpected field number %v of implicit list struct", fnum)
					return
				}
				// Read field value of implicit struct.
				efopts := fopts
				efopts.BinFieldNum = 0 // dontcare
				_n, err = cdc.decodeReflectBinary(ibz, einfo, erv, efopts, false, 0)
				if slide(&ibz, &n, _n) && err != nil {
					err = fmt.Errorf("error reading array contents: %w", err)
					return
				}
				// Ensure that there are no more bytes left.
				if len(ibz) > 0 {
					err = fmt.Errorf("unexpected trailing bytes after implicit list struct's Value field: %X", ibz)
					return
				}
			} else {
				// General case
				efopts := fopts
				efopts.BinFieldNum = 1
				_n, err = cdc.decodeReflectBinary(bz, einfo, erv, efopts, false, 0)
				if slide(&bz, &n, _n) && err != nil {
					err = fmt.Errorf("error reading array contents: %w", err)
					return
				}
			}
		}
		// Ensure that there are no more elements left,
		// and no field number regression either.
		// This is to provide better error messages.
		if len(bz) > 0 {
			var fnum uint32
			fnum, _, _, err = decodeFieldNumberAndTyp3(bz)
			if err != nil {
				return
			}
			if fnum <= fopts.BinFieldNum {
				err = fmt.Errorf("unexpected field number %v after repeated field number %v", fnum, fopts.BinFieldNum)
				return
			}
		}
	}
	return n, err
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryByteSlice(bz []byte, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions,
) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectByteSlice")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}
	// If len(bz) == 0 the code below will err
	if len(bz) == 0 {
		rv.Set(info.ZeroValue)
		return 0, nil
	}

	// Read byte-length prefixed byteslice.
	var (
		byteslice []byte
		_n        int
	)
	byteslice, _n, err = DecodeByteSlice(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if len(byteslice) == 0 {
		// Special case when length is 0.
		// NOTE: We prefer nil slices.
		rv.Set(info.ZeroValue)
	} else {
		rv.Set(reflect.ValueOf(byteslice))
	}
	return n, err
}

// CONTRACT: rv.CanAddr() is true.
// NOTE: Keep the code structure similar to decodeReflectBinaryArray.
func (cdc *Codec) decodeReflectBinarySlice(bz []byte, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions, bare bool,
) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectBinarySlice")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}
	einfo, err := cdc.getTypeInfoWLock(ert)
	if err != nil {
		return
	}

	// Construct slice to collect decoded items to.
	// NOTE: This is due to Proto3.  How to best optimize?
	esrt := reflect.SliceOf(ert)
	srv := reflect.Zero(esrt)

	// Strip if needed.
	bz, err = decodeMaybeBare(bz, &n, bare)
	if err != nil {
		return
	}

	// If elem is not already a ByteLength type, read in packed form.
	// This is a Proto wart due to Proto backwards compatibility issues.
	// Amino2 will probably migrate to use the List typ3.
	newoptions := uint64(0)
	// Special case for list of (repr) bytes: encode as "bytes".
	if einfo.ReprType.Type.Kind() == reflect.Uint8 {
		newoptions |= beOptionByte
	}
	typ3 := einfo.GetTyp3(fopts)
	if typ3 != Typ3ByteLength || (newoptions&beOptionByte > 0) {
		// Read elems in packed form.
		for len(bz) != 0 {
			erv, _n := reflect.New(ert).Elem(), int(0)
			_n, err = cdc.decodeReflectBinary(bz, einfo, erv, fopts, false, newoptions)
			if slide(&bz, &n, _n) && err != nil {
				err = fmt.Errorf("error reading array contents: %w", err)
				return
			}
			srv = reflect.Append(srv, erv)
		}
	} else {
		// NOTE: ert is for the element value, while einfo.Type is dereferenced.
		isErtStructPointer := ert.Kind() == reflect.Ptr && einfo.Type.Kind() == reflect.Struct
		writeImplicit := isListType(einfo.Type) &&
			einfo.Elem.ReprType.Type.Kind() != reflect.Uint8 &&
			einfo.Elem.ReprType.GetTyp3(fopts) != Typ3ByteLength

		// Read elements in unpacked form.
		for len(bz) != 0 {
			// Read field key (number and type).
			var (
				typ  Typ3
				_n   int
				fnum uint32
			)
			fnum, typ, _n, err = decodeFieldNumberAndTyp3(bz)
			if fnum > fopts.BinFieldNum {
				break // before sliding...
			}
			if slide(&bz, &n, _n) && err != nil {
				return
			}
			// Validate field number and typ3.
			if fnum < fopts.BinFieldNum {
				err = errors.New(fmt.Sprintf("expected repeated field number %v or greater, got %v", fopts.BinFieldNum, fnum))
				return
			}
			if typ != Typ3ByteLength {
				err = errors.New(fmt.Sprintf("expected repeated field type %v, got %v", Typ3ByteLength, typ))
				return
			}
			// Decode the next ByteLength bytes into erv.
			erv, _n := reflect.New(ert).Elem(), int(0)
			// Special case if:
			//  * next ByteLength bytes are 0x00, and
			//  * - erv is not a struct pointer, or
			//    - field option has NilElements set
			if (len(bz) > 0 && bz[0] == 0x00) &&
				(!isErtStructPointer || fopts.NilElements) {
				slide(&bz, &n, 1)
				erv.Set(defaultValue(erv.Type()))
				srv = reflect.Append(srv, erv)
				continue
			}
			// Special case: nested lists.
			// Multidimensional lists (nested inner lists also in unpacked
			// form) are represented as lists of implicit structs.
			if writeImplicit {
				// Read bytes for implicit struct.
				var ibz []byte
				ibz, _n, err = DecodeByteSlice(bz)
				if slide(&bz, nil, _n) && err != nil {
					return
				}
				// This is a trick for debuggability -- we slide on &n more later.
				n += UvarintSize(uint64(len(ibz)))
				// Read field key of implicit struct.
				var fnum uint32
				fnum, _, _n, err = decodeFieldNumberAndTyp3(ibz)
				if slide(&ibz, &n, _n) && err != nil {
					return
				}
				if fnum != 1 {
					err = fmt.Errorf("unexpected field number %v of implicit list struct", fnum)
					return
				}
				// Read field value of implicit struct.
				efopts := fopts
				efopts.BinFieldNum = 0 // dontcare
				_n, err = cdc.decodeReflectBinary(ibz, einfo, erv, efopts, false, 0)
				if slide(&ibz, &n, _n) && err != nil {
					err = fmt.Errorf("error reading slice contents: %w", err)
					return
				}
				// Ensure that there are no more bytes left.
				if len(ibz) > 0 {
					err = fmt.Errorf("unexpected trailing bytes after implicit list struct's Value field: %X", ibz)
					return
				}
			} else {
				// General case
				efopts := fopts
				efopts.BinFieldNum = 1
				_n, err = cdc.decodeReflectBinary(bz, einfo, erv, efopts, false, 0)
				if slide(&bz, &n, _n) && err != nil {
					err = fmt.Errorf("error reading slice contents: %w", err)
					return
				}
			}
			srv = reflect.Append(srv, erv)
		}
	}
	rv.Set(srv)
	return n, err
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectBinaryStruct(bz []byte, info *TypeInfo, rv reflect.Value,
	_ FieldOptions, bare bool,
) (n int, err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectBinaryStruct")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}
	_n := 0 //nolint: ineffassign

	// NOTE: The "Struct" typ3 doesn't get read here.
	// It's already implied, either by struct-key or list-element-type-byte.

	// Strip if needed.
	bz, err = decodeMaybeBare(bz, &n, bare)
	if err != nil {
		return
	}

	// Track the last seen field number.
	var lastFieldNum uint32
	// Read each field.
	for _, field := range info.Fields {
		// Get field rv and info.
		frv := rv.Field(field.Index)
		finfo := field.TypeInfo

		// We're done if we've consumed all of bz.
		if len(bz) == 0 {
			frv.Set(defaultValue(frv.Type()))
			continue
		}

		if field.UnpackedList {
			// Skip unpacked list field if fnum is bigger.
			var fnum uint32
			fnum, _, _, err = decodeFieldNumberAndTyp3(bz)
			if err != nil {
				return
			}
			if field.BinFieldNum < fnum {
				continue
			}
			// This is a list that was encoded unpacked, e.g.
			// with repeated field entries for each list item.
			_n, err = cdc.decodeReflectBinary(bz, finfo, frv, field.FieldOptions, true, 0)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
		} else {
			// Read field key (number and type).
			var (
				fnum uint32
				typ  Typ3
			)
			fnum, typ, _n, err = decodeFieldNumberAndTyp3(bz)
			if field.BinFieldNum < fnum {
				// Set zero field value.
				frv.Set(defaultValue(frv.Type()))
				continue // before sliding...
			}
			if slide(&bz, &n, _n) && err != nil {
				return
			}

			// Validate fnum and typ.
			if fnum <= lastFieldNum {
				err = fmt.Errorf("encountered fieldNum: %v, but we have already seen fnum: %v\nbytes:%X",
					fnum, lastFieldNum, bz)
				return
			}
			lastFieldNum = fnum
			// NOTE: In the future, we'll support upgradeability.
			// So in the future, this may not match,
			// so we will need to remove this sanity check.
			if field.BinFieldNum != fnum {
				err = errors.New(fmt.Sprintf("expected field # %v of %v, got %v",
					field.BinFieldNum, info.Type, fnum))
				return
			}
			typWanted := finfo.GetTyp3(field.FieldOptions)
			if typ != typWanted {
				err = errors.New(fmt.Sprintf("expected field type %v for # %v of %v, got %v",
					typWanted, fnum, info.Type, typ))
				return
			}
			// Decode field into frv.
			_n, err = cdc.decodeReflectBinary(bz, finfo, frv, field.FieldOptions, false, 0)
			if slide(&bz, &n, _n) && err != nil {
				return
			}
		}
	}

	// Consume any remaining fields.
	var (
		fnum uint32
		typ3 Typ3
	)
	for len(bz) > 0 {
		fnum, typ3, _n, err = decodeFieldNumberAndTyp3(bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
		if fnum <= lastFieldNum {
			err = fmt.Errorf("encountered fieldNum: %v, but we have already seen fnum: %v\nbytes:%X",
				fnum, lastFieldNum, bz)
			return
		}
		lastFieldNum = fnum

		_n, err = consumeAny(typ3, bz)
		if slide(&bz, &n, _n) && err != nil {
			return
		}
	}
	return n, err
}

// ----------------------------------------
// consume* for skipping struct fields

// Read everything without doing anything with it. Report errors if they occur.
func consumeAny(typ3 Typ3, bz []byte) (n int, err error) {
	var _n int
	switch typ3 {
	case Typ3Varint:
		_, _n, err = DecodeVarint(bz)
	case Typ38Byte:
		_, _n, err = DecodeInt64(bz)
	case Typ3ByteLength:
		_, _n, err = DecodeByteSlice(bz)
	case Typ34Byte:
		_, _n, err = DecodeInt32(bz)
	default:
		err = fmt.Errorf("invalid typ3 bytes %v", typ3)
		return
	}
	if err != nil {
		// do not slide
		return
	}
	slide(&bz, &n, _n)
	return
}

// ----------------------------------------

// Read field key.
func decodeFieldNumberAndTyp3(bz []byte) (num uint32, typ Typ3, n int, err error) {
	// Read uvarint value.
	value64, n, err := DecodeUvarint(bz)
	if err != nil {
		return
	}

	// Decode first typ3 byte.
	typ = Typ3(value64 & 0x07)

	// Decode num.
	num64 := value64 >> 3
	if num64 > (1<<29 - 1) {
		err = fmt.Errorf("invalid field num %v", num64)
		return
	}
	num = uint32(num64)
	return
}

// ----------------------------------------
// Misc.

func decodeMaybeBare(bz []byte, n *int, bare bool) ([]byte, error) {
	if bare {
		return bz, nil
	} else {
		// Read byte-length prefixed byteslice.
		var (
			buf []byte
			_n  int
			err error
		)
		buf, _n, err = DecodeByteSlice(bz)
		if slide(&bz, nil, _n) && err != nil {
			return bz, err
		}
		// This is a trick for debuggability -- we slide on &n more later.
		*n += UvarintSize(uint64(len(buf)))
		bz = buf
		return bz, nil
	}
}

package amino

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

// ----------------------------------------
// cdc.encodeReflectJSON

// This is the main entrypoint for encoding all types in json form.  This
// function calls encodeReflectJSON*, and generally those functions should
// only call this one, for the disfix wrapper is only written here.
// NOTE: Unlike encodeReflectBinary, rv may be a pointer.  This is because
// unlike the binary representation, in JSON there is a concrete representation
// of no value -- null.  So, a nil pointer here encodes as null, whereas
// encodeReflectBinary() assumes that the pointer is already dereferenced.
// CONTRACT: rv is valid.
func (cdc *Codec) encodeReflectJSON(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (err error) {
	if !rv.IsValid() {
		panic("should not happen")
	}
	if printLog {
		fmt.Printf("(E) encodeReflectJSON(info: %v, rv: %#v (%v), fopts: %v)\n",
			info, rv.Interface(), rv.Type(), fopts)
		defer func() {
			fmt.Printf("(E) -> err: %v\n", err)
		}()
	}

	// Dereference value if pointer.
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			err = writeStr(w, `null`)
			return
		}
		rv = rv.Elem()
	}

	// Handle the most special case, "well known".
	if info.IsJSONWellKnownType {
		var ok bool
		ok, err = encodeReflectJSONWellKnown(w, info, rv, fopts)
		if ok || err != nil {
			return
		}
	}

	// Handle override if rv implements amino.Marshaler.
	if info.IsAminoMarshaler {
		// First, encode rv into repr instance.
		var (
			rrv   reflect.Value
			rinfo *TypeInfo
		)
		rrv, err = toReprObject(rv)
		if err != nil {
			return
		}
		rinfo = info.ReprType
		// Then, encode the repr instance.
		err = cdc.encodeReflectJSON(w, rinfo, rrv, fopts)
		return
	}

	switch info.Type.Kind() {
	// ----------------------------------------
	// Complex

	case reflect.Interface:
		return cdc.encodeReflectJSONInterface(w, info, rv, fopts)

	case reflect.Array, reflect.Slice:
		return cdc.encodeReflectJSONList(w, info, rv, fopts)

	case reflect.Struct:
		return cdc.encodeReflectJSONStruct(w, info, rv, fopts)

	// ----------------------------------------
	// Signed, Unsigned

	case reflect.Int64, reflect.Int:
		_, err = fmt.Fprintf(w, `"%d"`, rv.Int()) // JS can't handle int64
		return

	case reflect.Uint64, reflect.Uint:
		_, err = fmt.Fprintf(w, `"%d"`, rv.Uint()) // JS can't handle uint64
		return

	case reflect.Int32, reflect.Int16, reflect.Int8,
		reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return invokeStdlibJSONMarshal(w, rv.Interface())

	// ----------------------------------------
	// Misc

	case reflect.Float64, reflect.Float32:
		if !fopts.Unsafe {
			return errors.New("amino.JSON float* support requires `amino:\"unsafe\"`")
		}
		fallthrough
	case reflect.Bool, reflect.String:
		return invokeStdlibJSONMarshal(w, rv.Interface())

	// ----------------------------------------
	// Default

	default:
		panic(fmt.Sprintf("unsupported type %v", info.Type.Kind()))
	}
}

func (cdc *Codec) encodeReflectJSONInterface(w io.Writer, iinfo *TypeInfo, rv reflect.Value,
	fopts FieldOptions,
) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectJSONInterface")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Special case when rv is nil, just write "null".
	if rv.IsNil() {
		err = writeStr(w, `null`)
		return
	}

	// Get concrete non-pointer reflect value & type.
	crv := rv.Elem()
	_, crvIsPtr, crvIsNilPtr := maybeDerefValue(crv)
	if crvIsPtr && crv.Kind() == reflect.Interface {
		// See "MARKER: No interface-pointers" in codec.go
		panic("should not happen")
	}
	if crvIsNilPtr {
		panic(fmt.Sprintf("Illegal nil-pointer of type %v for registered interface %v. "+
			"For compatibility with other languages, nil-pointer interface values are forbidden.", crv.Type(), iinfo.Type))
	}
	crt := crv.Type()

	// Get *TypeInfo for concrete type.
	var cinfo *TypeInfo
	cinfo, err = cdc.getTypeInfoWLock(crt)
	if err != nil {
		return
	}
	if !cinfo.Registered {
		err = errors.New("cannot encode unregistered concrete type %v", crt)
		return
	}

	// Write Value to buffer
	buf := new(bytes.Buffer)
	cdc.encodeReflectJSON(buf, cinfo, crv, fopts)
	value := buf.Bytes()
	if len(value) == 0 {
		err = errors.New("JSON bytes cannot be empty")
		return
	}
	if cinfo.IsJSONAnyValueType || (cinfo.IsAminoMarshaler && cinfo.ReprType.IsJSONAnyValueType) {
		// Sanity check
		if value[0] == '{' || value[len(value)-1] == '}' {
			err = errors.New("unexpected JSON object %s", value)
			return
		}
		// Write TypeURL
		err = writeStr(w, _fmt(`{"@type":"%s","value":`, cinfo.TypeURL))
		if err != nil {
			return
		}
		// Write Value
		err = writeStr(w, string(value))
		if err != nil {
			return
		}
		// Write closing brace.
		err = writeStr(w, `}`)
		return
	} else {
		// Sanity check
		if value[0] != '{' || value[len(value)-1] != '}' {
			err = errors.New("expected JSON object but got %s", value)
			return
		}
		// Write TypeURL
		err = writeStr(w, _fmt(`{"@type":"%s"`, cinfo.TypeURL))
		if err != nil {
			return
		}
		// Write Value
		if len(value) > 2 {
			err = writeStr(w, ","+string(value[1:]))
		} else {
			err = writeStr(w, `}`)
		}
		return
	}
}

func (cdc *Codec) encodeReflectJSONList(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectJSONList")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Special case when list is a nil slice, just write "null".
	// Empty slices and arrays are not encoded as "null".
	if rv.Kind() == reflect.Slice && rv.IsNil() {
		err = writeStr(w, `null`)
		return
	}

	ert := info.Type.Elem()
	length := rv.Len()

	switch ert.Kind() {
	case reflect.Uint8: // Special case: byte array
		// Write bytes in base64.
		// NOTE: Base64 encoding preserves the exact original number of bytes.
		// Get readable slice of bytes.
		var bz []byte
		if rv.CanAddr() {
			bz = rv.Slice(0, length).Bytes()
		} else {
			bz = make([]byte, length)
			reflect.Copy(reflect.ValueOf(bz), rv) // XXX: looks expensive!
		}
		var jsonBytes []byte
		jsonBytes, err = json.Marshal(bz) // base64 encode
		if err != nil {
			return
		}
		_, err = w.Write(jsonBytes)
		return

	default:
		// Open square bracket.
		err = writeStr(w, `[`)
		if err != nil {
			return
		}

		// Write elements with comma.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfoWLock(ert)
		if err != nil {
			return
		}
		for i := 0; i < length; i++ {
			// Get dereferenced element value and info.
			erv := rv.Index(i)
			if erv.Kind() == reflect.Ptr &&
				erv.IsNil() {
				// then
				err = writeStr(w, `null`)
			} else {
				err = cdc.encodeReflectJSON(w, einfo, erv, fopts)
			}
			if err != nil {
				return
			}
			// Add a comma if it isn't the last item.
			if i != length-1 {
				err = writeStr(w, `,`)
				if err != nil {
					return
				}
			}
		}

		// Close square bracket.
		defer func() {
			err = writeStr(w, `]`)
		}()
		return
	}
}

func (cdc *Codec) encodeReflectJSONStruct(w io.Writer, info *TypeInfo, rv reflect.Value, _ FieldOptions) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectJSONStruct")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Part 1.
	err = writeStr(w, `{`)
	if err != nil {
		return
	}
	// Part 2.
	defer func() {
		if err == nil {
			err = writeStr(w, `}`)
		}
	}()

	writeComma := false
	for _, field := range info.Fields {
		finfo := field.TypeInfo
		// Get dereferenced field value and info.
		frv, _, frvIsNil := maybeDerefValue(rv.Field(field.Index))
		// If frv is empty and omitempty, skip it.
		// NOTE: Unlike Amino:binary, we don't skip null fields unless "omitempty".
		if field.JSONOmitEmpty && isJSONEmpty(frv, field.ZeroValue) {
			continue
		}
		// Now we know we're going to write something.
		// Add a comma if we need to.
		if writeComma {
			err = writeStr(w, `,`)
			if err != nil {
				return
			}
			writeComma = false //nolint:ineffassign
		}
		// Write field JSON name.
		err = invokeStdlibJSONMarshal(w, field.JSONName)
		if err != nil {
			return
		}
		// Write colon.
		err = writeStr(w, `:`)
		if err != nil {
			return
		}
		// Write field value.
		if frvIsNil {
			err = writeStr(w, `null`)
		} else {
			err = cdc.encodeReflectJSON(w, finfo, frv, field.FieldOptions)
		}
		if err != nil {
			return
		}
		writeComma = true
	}
	return err
}

// ----------------------------------------
// Misc.

func invokeStdlibJSONMarshal(w io.Writer, v interface{}) error {
	// Note: Please don't stream out the output because that adds a newline
	// using json.NewEncoder(w).Encode(data)
	// as per https://golang.org/pkg/encoding/json/#Encoder.Encode
	blob, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(blob)
	return err
}

func writeStr(w io.Writer, s string) (err error) {
	_, err = w.Write([]byte(s))
	return
}

func _fmt(s string, args ...interface{}) string {
	return fmt.Sprintf(s, args...)
}

// For json:",omitempty".
// Returns true for zero values, but also non-nil zero-length slices and strings.
func isJSONEmpty(rv reflect.Value, zrv reflect.Value) bool {
	if !rv.IsValid() {
		return true
	}
	if reflect.DeepEqual(rv.Interface(), zrv.Interface()) {
		return true
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		if rv.Len() == 0 {
			return true
		}
	}
	return false
}

func isJSONAnyValueType(rt reflect.Type) bool {
	if isJSONWellKnownType(rt) {
		// All well known types are to be encoded as "{@type,value}" in
		// JSON.  Some of these may be structs/objects, such as
		// gAnyType, but nevertheless they must be encoded as
		// {@type,value}, the latter specifically
		// {@type:"/google.protobuf.Any",value:{@type,value}).
		return true
	}
	// Otherwise, it depends on the kind.
	switch rt.Kind() {
	case
		reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		// Primitive types get special {@type,value} treatment.  In
		// binary form, most of these types would be encoded
		// wrapped in an implicit struct, except for lists (both of
		// bytes and of anything else), and for strings...
		reflect.Array, reflect.Slice, reflect.String:
		// ...which are all non-objects that must be encoded as
		// {@type,value}.
		return true
	default:
		return false
	}
}

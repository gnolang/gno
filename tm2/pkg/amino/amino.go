package amino

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// Package "pkg" exists So dependencies can create Packages.
// We export it here so this amino package can use it natively.
type (
	Package = pkg.Package
	Type    = pkg.Type
)

var (
	// Global methods for global auto-sealing codec.
	gcdc *Codec

	// we use this time to init. an empty value (opposed to reflect.Zero which gives time.Time{} / 01-01-01 00:00:00)
	emptyTime time.Time

	// ErrNoPointer is thrown when you call a method that expects a pointer, e.g. Unmarshal
	ErrNoPointer = errors.New("expected a pointer")
)

const (
	unixEpochStr = "1970-01-01 00:00:00 +0000 UTC"
	epochFmt     = "2006-01-02 15:04:05 +0000 UTC"
)

func init() {
	gcdc = NewCodec().WithPBBindings().Autoseal()
	var err error
	emptyTime, err = time.Parse(epochFmt, unixEpochStr)
	if err != nil {
		panic("couldn't parse empty value for time")
	}
}

// XXX reorder global and cdc methods for consistency and logic.

func Marshal(o any) ([]byte, error) {
	return gcdc.Marshal(o)
}

func MustMarshal(o any) []byte {
	return gcdc.MustMarshal(o)
}

func MarshalSized(o any) ([]byte, error) {
	return gcdc.MarshalSized(o)
}

func MarshalSizedWriter(w io.Writer, o any) (n int64, err error) {
	return gcdc.MarshalSizedWriter(w, o)
}

func MustMarshalSized(o any) []byte {
	return gcdc.MustMarshalSized(o)
}

func MarshalAny(o any) ([]byte, error) {
	return gcdc.MarshalAny(o)
}

func MustMarshalAny(o any) []byte {
	return gcdc.MustMarshalAny(o)
}

func MarshalAnySized(o any) ([]byte, error) {
	return gcdc.MarshalAnySized(o)
}

func MustMarshalAnySized(o any) []byte {
	return gcdc.MustMarshalAnySized(o)
}

func MarshalAnySizedWriter(w io.Writer, o any) (n int64, err error) {
	return gcdc.MarshalAnySizedWriter(w, o)
}

func Unmarshal(bz []byte, ptr any) error {
	return gcdc.Unmarshal(bz, ptr)
}

func MustUnmarshal(bz []byte, ptr any) {
	gcdc.MustUnmarshal(bz, ptr)
}

func UnmarshalSized(bz []byte, ptr any) error {
	return gcdc.UnmarshalSized(bz, ptr)
}

func UnmarshalSizedReader(r io.Reader, ptr any, maxSize int64) (n int64, err error) {
	return gcdc.UnmarshalSizedReader(r, ptr, maxSize)
}

func MustUnmarshalSized(bz []byte, ptr any) {
	gcdc.MustUnmarshalSized(bz, ptr)
}

func UnmarshalAny(bz []byte, ptr any) error {
	return gcdc.UnmarshalAny(bz, ptr)
}

func UnmarshalAny2(typeURL string, value []byte, ptr any) error {
	return gcdc.UnmarshalAny2(typeURL, value, ptr)
}

func MustUnmarshalAny(bz []byte, ptr any) {
	gcdc.MustUnmarshalAny(bz, ptr)
}

func UnmarshalAnySized(bz []byte, ptr any) error {
	return gcdc.UnmarshalAnySized(bz, ptr)
}

func MarshalJSON(o any) ([]byte, error) {
	return gcdc.JSONMarshal(o)
}

func MarshalJSONAny(o any) ([]byte, error) {
	return gcdc.MarshalJSONAny(o)
}

func MustMarshalJSON(o any) []byte {
	return gcdc.MustMarshalJSON(o)
}

func MustMarshalJSONAny(o any) []byte {
	return gcdc.MustMarshalJSONAny(o)
}

func UnmarshalJSON(bz []byte, ptr any) error {
	return gcdc.JSONUnmarshal(bz, ptr)
}

func MustUnmarshalJSON(bz []byte, ptr any) {
	gcdc.MustUnmarshalJSON(bz, ptr)
}

func MarshalJSONIndent(o any, prefix, indent string) ([]byte, error) {
	return gcdc.MarshalJSONIndent(o, prefix, indent)
}

// XXX unstable API.
func GetTypeURL(o any) string {
	return gcdc.GetTypeURL(o)
}

// Returns a new TypeInfo instance.
// NOTE: it uses a new codec for security's sake.
// (*TypeInfo of gcdc should not be exposed)
// Therefore it may be inefficient.  If you need efficiency, implement with a
// new method that takes as argument a non-global codec instance.
func GetTypeInfo(rt reflect.Type) (info *TypeInfo, err error) {
	cdc := NewCodec().WithPBBindings().Autoseal()
	ti, err := cdc.GetTypeInfo(rt)
	return ti, err
}

// ----------------------------------------
// Typ3

type Typ3 uint8

const (
	// Typ3 types
	Typ3Varint     = Typ3(0)
	Typ38Byte      = Typ3(1)
	Typ3ByteLength = Typ3(2)
	// Typ3_Struct     = Typ3(3)
	// Typ3_StructTerm = Typ3(4)
	Typ34Byte = Typ3(5)
	// Typ3_List       = Typ3(6)
	// Typ3_Interface  = Typ3(7)
)

func (typ Typ3) String() string {
	switch typ {
	case Typ3Varint:
		return "(U)Varint"
	case Typ38Byte:
		return "8Byte"
	case Typ3ByteLength:
		return "ByteLength"
	// case Typ3_Struct:
	//	return "Struct"
	// case Typ3_StructTerm:
	//	return "StructTerm"
	case Typ34Byte:
		return "4Byte"
	// case Typ3_List:
	//	return "List"
	// case Typ3_Interface:
	//	return "Interface"
	default:
		return fmt.Sprintf("<Invalid Typ3 %X>", byte(typ))
	}
}

// ----------------------------------------
// *Codec methods

// ----------------------------------------
// Marshal* methods

// MarshalSized encodes the object o according to the Amino spec,
// but prefixed by a uvarint encoding of the object to encode.
// Use Marshal if you don't want byte-length prefixing.
//
// For consistency, MarshalSized will first dereference pointers
// before encoding.  MarshalSized will panic if o is a nil-pointer,
// or if o is invalid.
func (cdc *Codec) MarshalSized(o any) ([]byte, error) {
	cdc.doAutoseal()

	// Write the bytes here.
	buf := poolBytesBuffer.Get()
	defer poolBytesBuffer.Put(buf)

	// Write the bz without length-prefixing.
	bz, err := cdc.Marshal(o)
	if err != nil {
		return nil, err
	}

	// Write uvarint(len(bz)).
	err = EncodeUvarint(buf, uint64(len(bz)))
	if err != nil {
		return nil, err
	}

	// Write bz.
	_, err = buf.Write(bz)
	if err != nil {
		return nil, err
	}

	return copyBytes(buf.Bytes()), nil
}

// MarshalSizedWriter writes the bytes as would be returned from
// MarshalSized to the writer w.
func (cdc *Codec) MarshalSizedWriter(w io.Writer, o any) (n int64, err error) {
	var (
		bz []byte
		_n int
	)
	bz, err = cdc.MarshalSized(o)
	if err != nil {
		return 0, err
	}
	_n, err = w.Write(bz) // TODO: handle overflow in 32-bit systems.
	n = int64(_n)
	return
}

// Panics if error.
func (cdc *Codec) MustMarshalSized(o any) []byte {
	bz, err := cdc.MarshalSized(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *Codec) MarshalAnySized(o any) ([]byte, error) {
	cdc.doAutoseal()

	// Write the bytes here.
	buf := poolBytesBuffer.Get()
	defer poolBytesBuffer.Put(buf)
	// Write the bz without length-prefixing.
	bz, err := cdc.MarshalAny(o)
	if err != nil {
		return nil, err
	}

	// Write uvarint(len(bz)).
	err = EncodeUvarint(buf, uint64(len(bz)))
	if err != nil {
		return nil, err
	}

	// Write bz.
	_, err = buf.Write(bz)
	if err != nil {
		return nil, err
	}

	return copyBytes(buf.Bytes()), nil
}

func (cdc *Codec) MustMarshalAnySized(o any) []byte {
	bz, err := cdc.MarshalAnySized(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *Codec) MarshalAnySizedWriter(w io.Writer, o any) (n int64, err error) {
	var (
		bz []byte
		_n int
	)
	bz, err = cdc.MarshalAnySized(o)
	if err != nil {
		return 0, err
	}
	_n, err = w.Write(bz) // TODO: handle overflow in 32-bit systems.
	n = int64(_n)
	return
}

// Marshal encodes the object o according to the Amino spec.
// Marshal doesn't prefix the byte-length of the encoding,
// so the caller must handle framing.
// Type information as in google.protobuf.Any isn't included, so manually wrap
// before calling if you need to decode into an interface.
// NOTE: nil-struct-pointers have no encoding. In the context of a struct,
// the absence of a field does denote a nil-struct-pointer, but in general
// this is not the case, so unlike MarshalJSON.
func (cdc *Codec) Marshal(o any) ([]byte, error) {
	cdc.doAutoseal()

	if cdc.usePBBindings {
		pbm, ok := o.(PBMessager)
		if ok {
			return cdc.MarshalPBBindings(pbm)
		}
		// Else, fall back to using reflection for native primitive types.
	}

	return cdc.MarshalReflect(o)
}

// Use reflection.
func (cdc *Codec) MarshalReflect(o any) ([]byte, error) {
	// Dereference value if pointer.
	rv := reflect.ValueOf(o)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			panic("Marshal cannot marshal a nil pointer directly. Try wrapping in a struct?")
			// NOTE: You can still do so by calling
			// `.MarshalSized(struct{ *SomeType })` or so on.
		}
		rv = rv.Elem()
		if rv.Kind() == reflect.Ptr {
			panic("nested pointers not allowed")
		}
	}

	// Encode Amino:binary bytes.
	var bz []byte
	buf := poolBytesBuffer.Get()
	defer poolBytesBuffer.Put(buf)

	rt := rv.Type()
	info, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return nil, err
	}
	// Implicit struct or not?
	// NOTE: similar to binary interface encoding.
	fopts := FieldOptions{}
	if !info.IsStructOrUnpacked(fopts) {
		writeEmpty := false
		// Encode with an implicit struct, with a single field with number 1.
		// The type of this implicit field determines whether any
		// length-prefixing happens after the typ3 byte.
		// The second FieldOptions is empty, because this isn't a list of
		// Typ3_ByteLength things, so however it is encoded, that option is no
		// longer needed.
		if err = cdc.writeFieldIfNotEmpty(buf, 1, info, FieldOptions{}, FieldOptions{}, rv, writeEmpty); err != nil {
			return nil, err
		}
		bz = copyBytes(buf.Bytes())
	} else {
		// The passed in BinFieldNum is only relevant for when the type is to
		// be encoded unpacked (elements are Typ3_ByteLength).  In that case,
		// encodeReflectBinary will repeat the field number as set here, as if
		// encoded with an implicit struct.
		err = cdc.encodeReflectBinary(buf, info, rv, FieldOptions{BinFieldNum: 1}, true, 0)
		if err != nil {
			return nil, err
		}
		bz = copyBytes(buf.Bytes())
	}
	// If bz is empty, prefer nil.
	if len(bz) == 0 {
		bz = nil
	}
	return bz, nil
}

// Use pbbindings.
func (cdc *Codec) MarshalPBBindings(pbm PBMessager) ([]byte, error) {
	pbo, err := pbm.ToPBMessage(cdc)
	if err != nil {
		return nil, err
	}
	bz, err := proto.Marshal(pbo)
	return bz, err
}

// Panics if error.
func (cdc *Codec) MustMarshal(o any) []byte {
	bz, err := cdc.Marshal(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// MarshalAny encodes the registered object
// wrapped with google.protobuf.Any.
func (cdc *Codec) MarshalAny(o any) ([]byte, error) {
	cdc.doAutoseal()

	// o cannot be nil, otherwise we don't know what type it is.
	if o == nil {
		return nil, errors.New("MarshalAny() requires non-nil argument")
	}

	// Dereference value if pointer.
	rv, _, _ := maybeDerefValue(reflect.ValueOf(o))
	rt := rv.Type()

	// rv cannot be an interface.
	if rv.Kind() == reflect.Interface {
		return nil, errors.New("MarshalAny() requires registered concrete type")
	}

	// Make a temporary interface var, to contain the value of o.
	ivar := rv.Interface()
	var iinfo *TypeInfo
	iinfo, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return nil, err
	}

	// Encode as interface.
	buf := poolBytesBuffer.Get()
	defer poolBytesBuffer.Put(buf)
	err = cdc.encodeReflectBinaryInterface(buf, iinfo, reflect.ValueOf(&ivar).Elem(), FieldOptions{}, true)
	if err != nil {
		return nil, err
	}
	bz := copyBytes(buf.Bytes())

	return bz, nil
}

func copyBytes(bz []byte) []byte {
	cp := make([]byte, len(bz))
	copy(cp, bz)
	return cp
}

// Panics if error.
func (cdc *Codec) MustMarshalAny(o any) []byte {
	bz, err := cdc.MarshalAny(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// ----------------------------------------
// Unmarshal* methods

// Like Unmarshal, but will first decode the byte-length prefix.
// UnmarshalSized will panic if ptr is a nil-pointer.
// Returns an error if not all of bz is consumed.
func (cdc *Codec) UnmarshalSized(bz []byte, ptr any) error {
	if len(bz) == 0 {
		return errors.New("unmarshalSized cannot decode empty bytes")
	}

	// Read byte-length prefix.
	u64, n := binary.Uvarint(bz)
	if n < 0 {
		return errors.New("Error reading msg byte-length prefix: got code %v", n)
	}
	if u64 > uint64(len(bz)-n) {
		return errors.New("Not enough bytes to read in UnmarshalSized, want %v more bytes but only have %v",
			u64, len(bz)-n)
	} else if u64 < uint64(len(bz)-n) {
		return errors.New("Bytes left over in UnmarshalSized, should read %v more bytes but have %v",
			u64, len(bz)-n)
	}
	bz = bz[n:]

	// Decode.
	return cdc.Unmarshal(bz, ptr)
}

// Like Unmarshal, but will first read the byte-length prefix.
// UnmarshalSizedReader will panic if ptr is a nil-pointer.
// If maxSize is 0, there is no limit (not recommended).
func (cdc *Codec) UnmarshalSizedReader(r io.Reader, ptr any,
	maxSize int64,
) (n int64, err error) {
	if maxSize < 0 {
		panic("maxSize cannot be negative.")
	}

	// Read byte-length prefix.
	var l int64
	var buf [binary.MaxVarintLen64]byte
	for i := range len(buf) {
		_, err = r.Read(buf[i : i+1])
		if err != nil {
			return
		}
		n++
		if buf[i]&0x80 == 0 {
			break
		}
		if n >= maxSize {
			err = errors.New(
				"read overflow, maxSize is %v but uvarint(length-prefix) is itself greater than maxSize",
				maxSize,
			)
		}
	}
	u64, _ := binary.Uvarint(buf[:])
	if err != nil {
		return
	}
	if maxSize > 0 {
		if uint64(maxSize) < u64 {
			err = errors.New("read overflow, maxSize is %v but this amino binary object is %v bytes", maxSize, u64)
			return
		}
		if (maxSize - n) < int64(u64) {
			err = errors.New(
				"read overflow, maxSize is %v but this length-prefixed amino binary object is %v+%v bytes",
				maxSize, n, u64,
			)
			return
		}
	}
	l = int64(u64)
	if l < 0 {
		_ = errors.New( //nolint:errcheck
			"read overflow, this implementation can't read this because, why would anyone have this much data? Hello from 2018",
		)
	}

	// Read that many bytes.
	bz := make([]byte, l)
	_, err = io.ReadFull(r, bz)
	if err != nil {
		return
	}
	n += l

	// Decode.
	err = cdc.Unmarshal(bz, ptr)
	return n, err
}

// Panics if error.
func (cdc *Codec) MustUnmarshalSized(bz []byte, ptr any) {
	err := cdc.UnmarshalSized(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// Like UnmarshalAny, but will first decode the byte-length prefix.
func (cdc *Codec) UnmarshalAnySized(bz []byte, ptr any) error {
	if len(bz) == 0 {
		return errors.New("unmarshalSized cannot decode empty bytes")
	}

	// Read byte-length prefix.
	u64, n := binary.Uvarint(bz)
	if n < 0 {
		return errors.New("Error reading msg byte-length prefix: got code %v", n)
	}
	if u64 > uint64(len(bz)-n) {
		return errors.New("Not enough bytes to read in UnmarshalAnySized, want %v more bytes but only have %v",
			u64, len(bz)-n)
	} else if u64 < uint64(len(bz)-n) {
		return errors.New("Bytes left over in UnmarshalAnySized, should read %v more bytes but have %v",
			u64, len(bz)-n)
	}
	bz = bz[n:]

	// Decode.
	return cdc.UnmarshalAny(bz, ptr)
}

// Unmarshal will panic if ptr is a nil-pointer.
func (cdc *Codec) Unmarshal(bz []byte, ptr any) error {
	cdc.doAutoseal()

	if cdc.usePBBindings {
		pbm, ok := ptr.(PBMessager)
		if ok {
			return cdc.unmarshalPBBindings(bz, pbm)
		}
		// Else, fall back to using reflection for native primitive types.
	}

	return cdc.unmarshalReflect(bz, ptr)
}

// Use reflection.
func (cdc *Codec) unmarshalReflect(bz []byte, ptr any) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return ErrNoPointer
	}
	rv = rv.Elem()
	rt := rv.Type()
	info, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return err
	}

	// See if we need to read the typ3 encoding of an implicit struct.
	//
	// If the dest ptr is an interface, it is assumed that the object is
	// wrapped in a google.protobuf.Any object, so skip this step.
	//
	// See corresponding encoding message in this file, and also
	// binary-decode.
	bare := true
	var nWrap int
	if !info.IsStructOrUnpacked(FieldOptions{}) &&
		len(bz) > 0 &&
		(rv.Kind() != reflect.Interface) {
		var (
			fnum      uint32
			typ       Typ3
			nFnumTyp3 int
		)
		fnum, typ, nFnumTyp3, err = decodeFieldNumberAndTyp3(bz)
		if err != nil {
			return errors.Wrap(err, "could not decode field number and type")
		}
		if fnum != 1 {
			return fmt.Errorf("expected field number: 1; got: %v", fnum)
		}
		typWanted := info.GetTyp3(FieldOptions{})
		if typ != typWanted {
			return fmt.Errorf("expected field type %v for # %v of %v, got %v",
				typWanted, fnum, info.Type, typ)
		}

		slide(&bz, &nWrap, nFnumTyp3)
		// "bare" is ignored when primitive, byteslice, bytearray.
		// When typ3 != ByteLength, then typ3 is one of Typ3Varint, Typ38Byte,
		// Typ34Byte; and they are all primitive.
		bare = false
	}

	// Decode contents into rv.
	n, err := cdc.decodeReflectBinary(bz, info, rv, FieldOptions{BinFieldNum: 1}, bare, 0)
	if err != nil {
		return fmt.Errorf(
			"unmarshal to %v failed after %d bytes (%w): %X",
			info.Type,
			n+nWrap,
			err,
			bz,
		)
	}
	if n != len(bz) {
		return fmt.Errorf(
			"unmarshal to %v didn't read all bytes. Expected to read %v, only read %v: %X",
			info.Type,
			len(bz),
			n+nWrap,
			bz,
		)
	}

	return nil
}

// Use pbbindings.
func (cdc *Codec) unmarshalPBBindings(bz []byte, pbm PBMessager) error {
	pbo := pbm.EmptyPBMessage(cdc)
	err := proto.Unmarshal(bz, pbo)
	if err != nil {
		rt := reflect.TypeOf(pbm)
		info, err2 := cdc.getTypeInfoWLock(rt)
		if err2 != nil {
			return err2
		}
		return errors.New("unmarshal to %v failed: %v",
			info.Type, err)
	}
	err = pbm.FromPBMessage(cdc, pbo)
	if err != nil {
		rt := reflect.TypeOf(pbm)
		info, err2 := cdc.getTypeInfoWLock(rt)
		if err2 != nil {
			return err2
		}
		return errors.New("unmarshal to %v failed: %v",
			info.Type, err)
	}
	return nil
}

// Panics if error.
func (cdc *Codec) MustUnmarshal(bz []byte, ptr any) {
	err := cdc.Unmarshal(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// UnmarshalAny decodes the registered object
// from an Any.
func (cdc *Codec) UnmarshalAny(bz []byte, ptr any) (err error) {
	cdc.doAutoseal()

	// Dereference ptr which must be pointer to interface.
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return ErrNoPointer
	}
	rv = rv.Elem()

	// Get interface *TypeInfo.
	iinfo, err := cdc.getTypeInfoWLock(rv.Type())
	if err != nil {
		return err
	}

	_, err = cdc.decodeReflectBinaryInterface(bz, iinfo, rv, FieldOptions{}, true)
	return
}

// like UnmarshalAny() but with typeURL and value destructured.
func (cdc *Codec) UnmarshalAny2(typeURL string, value []byte, ptr any) (err error) {
	cdc.doAutoseal()

	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return ErrNoPointer
	}
	rv = rv.Elem()
	_, err = cdc.decodeReflectBinaryAny(typeURL, value, rv, FieldOptions{})
	return
}

func (cdc *Codec) MustUnmarshalAny(bz []byte, ptr any) {
	err := cdc.UnmarshalAny(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (cdc *Codec) JSONMarshal(o any) ([]byte, error) {
	cdc.doAutoseal()

	rv := reflect.ValueOf(o)
	if !rv.IsValid() {
		return []byte("null"), nil
	}
	rt := rv.Type()
	w := poolBytesBuffer.Get()
	defer poolBytesBuffer.Put(w)
	info, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return nil, err
	}
	if err = cdc.encodeReflectJSON(w, info, rv, FieldOptions{}); err != nil {
		return nil, err
	}

	return copyBytes(w.Bytes()), nil
}

func (cdc *Codec) MarshalJSONAny(o any) ([]byte, error) {
	// o cannot be nil, otherwise we don't know what type it is.
	if o == nil {
		return nil, errors.New("MarshalJSONAny() requires non-nil argument")
	}

	// Dereference value if pointer.
	rv := reflect.ValueOf(o)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()

	// rv cannot be an interface.
	if rv.Kind() == reflect.Interface {
		return nil, errors.New("MarshalJSONAny() requires registered concrete type")
	}

	// Make a temporary interface var, to contain the value of o.
	ivar := rv.Interface()
	var iinfo *TypeInfo
	iinfo, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return nil, err
	}

	// Encode as interface.
	buf := poolBytesBuffer.Get()
	defer poolBytesBuffer.Put(buf)

	err = cdc.encodeReflectJSONInterface(buf, iinfo, reflect.ValueOf(&ivar).Elem(), FieldOptions{})
	if err != nil {
		return nil, err
	}
	bz := copyBytes(buf.Bytes())

	return bz, nil
}

// MustMarshalJSON panics if an error occurs. Besides that behaves exactly like MarshalJSON.
func (cdc *Codec) MustMarshalJSON(o any) []byte {
	bz, err := cdc.JSONMarshal(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// MustMarshalJSONAny panics if an error occurs. Besides that behaves exactly like MarshalJSONAny.
func (cdc *Codec) MustMarshalJSONAny(o any) []byte {
	bz, err := cdc.MarshalJSONAny(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *Codec) JSONUnmarshal(bz []byte, ptr any) error {
	cdc.doAutoseal()
	if len(bz) == 0 {
		return errors.New("cannot decode empty bytes")
	}

	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		return errors.New("expected a pointer")
	}
	rv = rv.Elem()
	rt := rv.Type()
	info, err := cdc.getTypeInfoWLock(rt)
	if err != nil {
		return err
	}
	return cdc.decodeReflectJSON(bz, info, rv, FieldOptions{})
}

// MustUnmarshalJSON panics if an error occurs. Besides that behaves exactly like UnmarshalJSON.
func (cdc *Codec) MustUnmarshalJSON(bz []byte, ptr any) {
	if err := cdc.JSONUnmarshal(bz, ptr); err != nil {
		panic(err)
	}
}

// MarshalJSONIndent calls json.Indent on the output of cdc.MarshalJSON
// using the given prefix and indent string.
func (cdc *Codec) MarshalJSONIndent(o any, prefix, indent string) ([]byte, error) {
	bz, err := cdc.JSONMarshal(o)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := json.Indent(&out, bz, prefix, indent); err != nil {
		return nil, err
	}
	return copyBytes(out.Bytes()), nil
}

// ----------------------------------------
// Other

// Given amino package `pi`, register it with the global codec.
// NOTE: do not modify the result.
func RegisterPackage(pi *pkg.Package) *Package {
	gcdc.RegisterPackage(pi)
	return pi
}

// Create an unregistered amino package with args:
// - (gopkg string) The Go package path, e.g. "github.com/gnolang/gno/tm2/pkg/std"
// - (p3pkg string) The (shorter) Proto3 package path (no slashes), e.g. "std"
// - (dirname string) Package directory this is called from. Typical is to use `amino.GetCallersDirname()`
func NewPackage(gopkg string, p3pkg string, dirname string) *Package {
	return pkg.NewPackage(gopkg, p3pkg, dirname)
}

// Get caller's package directory.
// Implementation uses `filepath.Dir(runtime.Caller(1))`.
// NOTE: duplicated in pkg/pkg.go; given what it does and how,
// both are probably needed.
func GetCallersDirname() string {
	dirname := "" // derive from caller.
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		panic("could not get caller to derive caller's package directory")
	}
	dirname = filepath.Dir(filename)
	if filename == "" || dirname == "" {
		panic("could not derive caller's package directory")
	}
	if !path.IsAbs(dirname) {
		dirname = "" // if relative, assume from module and return empty string
	}
	return dirname
}

// ----------------------------------------
// Object

// All concrete types must implement the Object interface for genproto
// bindings.  They are generated automatically by genproto/bindings.go
type Object interface {
	GetTypeURL() string
}

// TODO: this does need the cdc receiver,
// as it should also work for non-pbbindings-optimized types.
// Returns the default type url for the given concrete type.
// NOTE: It must be fast, as it is used in pbbindings.
// XXX Unstable API.
func (cdc *Codec) GetTypeURL(o any) string {
	if obj, ok := o.(Object); ok {
		return obj.GetTypeURL()
	}
	switch o.(type) {
	case time.Time, *time.Time, *timestamppb.Timestamp:
		return "/google.protobuf.Timestamp"
	case time.Duration, *time.Duration, *durationpb.Duration:
		return "/google.protobuf.Duration"
	}
	rt := reflect.TypeOf(o)
	// Doesn't have .GetTypeURL() and isn't well known.
	// Do the slow thing (not relevant if pbbindings exists).
	info, err := cdc.GetTypeInfo(rt)
	if err != nil {
		panic(err)
	}
	if info.TypeURL == "" {
		panic("not yet supported")
	}
	return info.TypeURL
}

// ----------------------------------------

// Methods generated by genproto/bindings.go for faster encoding.
type PBMessager interface {
	ToPBMessage(*Codec) (proto.Message, error)
	EmptyPBMessage(*Codec) proto.Message
	FromPBMessage(*Codec, proto.Message) error
}

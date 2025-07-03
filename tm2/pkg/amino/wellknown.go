package amino

// NOTE: We must not depend on protubuf libraries for serialization.

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	//"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	// native
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
	// doubleType   = reflect.TypeOf(float64(0))
	// floatType    = reflect.TypeOf(float32(0))
	int64Type  = reflect.TypeOf(int64(0))
	uint64Type = reflect.TypeOf(uint64(0))
	int32Type  = reflect.TypeOf(int32(0))
	uint32Type = reflect.TypeOf(uint32(0))
	int16Type  = reflect.TypeOf(int16(0))
	uint16Type = reflect.TypeOf(uint16(0))
	int8Type   = reflect.TypeOf(int8(0))
	uint8Type  = reflect.TypeOf(uint8(0))
	boolType   = reflect.TypeOf(bool(false))
	stringType = reflect.TypeOf(string(""))
	bytesType  = reflect.TypeOf([]byte(nil))
	intType    = reflect.TypeOf(int(0))
	uintType   = reflect.TypeOf(uint(0))

	// google
	gAnyType       = reflect.TypeOf(anypb.Any{})
	gTimestampType = reflect.TypeOf(timestamppb.Timestamp{})
	gDurationType  = reflect.TypeOf(durationpb.Duration{})
	gEmptyType     = reflect.TypeOf(emptypb.Empty{})
	// gStructType    = reflect.TypeOf(structpb.Struct{}) MAP not yet supported
	// gValueType     = reflect.TypeOf(structpb.Value{})
	// gListType      = reflect.TypeOf(structpb.ListValue{})
	// gDoubleType    = reflect.TypeOf(wrapperspb.DoubleValue{})
	// gFloatType     = reflect.TypeOf(wrapperspb.FloatValue{})
	gInt64Type  = reflect.TypeOf(wrapperspb.Int64Value{})
	gUInt64Type = reflect.TypeOf(wrapperspb.UInt64Value{})
	gInt32Type  = reflect.TypeOf(wrapperspb.Int32Value{})
	gUInt32Type = reflect.TypeOf(wrapperspb.UInt32Value{})
	gBoolType   = reflect.TypeOf(wrapperspb.BoolValue{})
	gStringType = reflect.TypeOf(wrapperspb.StringValue{})
	gBytesType  = reflect.TypeOf(wrapperspb.BytesValue{})
)

var (
	nativePkg = NewPackage(
		"",
		"",
		"",
	).
		WithP3SchemaFile("").
		WithTypes(
			int64(0), uint64(0), int32(0), uint32(0), bool(false),
			string(""), []byte(nil), int(0), uint(0),
		)

	timePkg = NewPackage(
		"time",
		"",
		"",
	).
		WithP3SchemaFile("").
		WithP3GoPkgPath(""). // since conflicting p3 pkg paths.
		WithTypes(
			time.Now(),
			time.Duration(0),
		)

	gAnyPkg = NewPackage(
		"google.golang.org/protobuf/types/known/anypb",
		"google.protobuf",
		"",
	).
		WithP3ImportPath("google/protobuf/any.proto").
		WithP3SchemaFile("").
		WithTypes(&anypb.Any{})

	gTimestampPkg = NewPackage(
		"google.golang.org/protobuf/types/known/timestamppb",
		"google.protobuf",
		"",
	).
		WithP3ImportPath("google/protobuf/timestamp.proto").
		WithP3SchemaFile("").
		WithTypes(&timestamppb.Timestamp{})

	gDurationPkg = NewPackage(
		"google.golang.org/protobuf/types/known/durationpb",
		"google.protobuf",
		"",
	).
		WithP3ImportPath("google/protobuf/duration.proto").
		WithP3SchemaFile("").
		WithTypes(&durationpb.Duration{})

	gEmptyPkg = NewPackage(
		"google.golang.org/protobuf/types/known/emptypb",
		"google.protobuf",
		"",
	).
		WithP3ImportPath("google/protobuf/empty.proto").
		WithP3SchemaFile("").
		WithTypes(&emptypb.Empty{})

	gWrappersPkg = NewPackage(
		"google.golang.org/protobuf/types/known/wrapperspb",
		"google.protobuf",
		"",
	).
		WithP3ImportPath("google/protobuf/wrappers.proto").
		WithP3SchemaFile("").
		WithTypes(
			&wrapperspb.BoolValue{},
			&wrapperspb.BytesValue{},
			&wrapperspb.DoubleValue{},
			&wrapperspb.FloatValue{},
			&wrapperspb.Int32Value{},
			&wrapperspb.Int64Value{},
			&wrapperspb.StringValue{},
			&wrapperspb.UInt32Value{},
			&wrapperspb.UInt64Value{},
		)
)

func (cdc *Codec) registerWellKnownTypes() {
	register, preferNative := true, false
	ptr, noPtr := true, false
	// native not supported by protobuf
	cdc.registerType(nativePkg, uint16Type, "/amino.UInt16", noPtr, register) // XXX create them, and consider switching other types over.
	cdc.registerType(nativePkg, uint8Type, "/amino.UInt8", noPtr, register)
	cdc.registerType(nativePkg, int16Type, "/amino.Int16", noPtr, register)
	cdc.registerType(nativePkg, int8Type, "/amino.Int8", noPtr, register)
	// native
	cdc.registerType(timePkg, timeType, "/google.protobuf.Timestamp", noPtr, register)
	cdc.registerType(timePkg, durationType, "/google.protobuf.Duration", noPtr, register)
	cdc.registerType(nativePkg, int64Type, "/google.protobuf.Int64Value", noPtr, register)
	cdc.registerType(nativePkg, uint64Type, "/google.protobuf.UInt64Value", noPtr, register)
	cdc.registerType(nativePkg, int32Type, "/google.protobuf.Int32Value", noPtr, register)
	cdc.registerType(nativePkg, uint32Type, "/google.protobuf.UInt32Value", noPtr, register)
	cdc.registerType(nativePkg, boolType, "/google.protobuf.BoolValue", noPtr, register)
	cdc.registerType(nativePkg, stringType, "/google.protobuf.StringValue", noPtr, register)
	cdc.registerType(nativePkg, bytesType, "/google.protobuf.BytesValue", noPtr, register)
	cdc.registerType(nativePkg, intType, "/google.protobuf.Int64Value", noPtr, preferNative)
	cdc.registerType(nativePkg, uintType, "/google.protobuf.UInt64Value", noPtr, preferNative)
	// google
	cdc.registerType(gAnyPkg, gAnyType, "/google.protobuf.Any", ptr, register)
	cdc.registerType(gDurationPkg, gDurationType, "/google.protobuf.Duration", ptr, preferNative)
	cdc.registerType(gEmptyPkg, gEmptyType, "/google.protobuf.Empty", ptr, register)
	cdc.registerType(gTimestampPkg, gTimestampType, "/google.protobuf.Timestamp", ptr, preferNative)
	cdc.registerType(gWrappersPkg, gInt64Type, "/google.protobuf.Int64Value", ptr, preferNative)
	cdc.registerType(gWrappersPkg, gUInt64Type, "/google.protobuf.UInt64Value", ptr, preferNative)
	cdc.registerType(gWrappersPkg, gInt32Type, "/google.protobuf.Int32Value", ptr, preferNative)
	cdc.registerType(gWrappersPkg, gUInt32Type, "/google.protobuf.UInt32Value", ptr, preferNative)
	cdc.registerType(gWrappersPkg, gBoolType, "/google.protobuf.BoolValue", ptr, preferNative)
	cdc.registerType(gWrappersPkg, gStringType, "/google.protobuf.StringValue", ptr, preferNative)
	cdc.registerType(gWrappersPkg, gBytesType, "/google.protobuf.BytesValue", ptr, preferNative)
}

// These require special functions for encoding/decoding.
func isBinaryWellKnownType(rt reflect.Type) (wellKnown bool) {
	switch rt {
	// Native types.
	case timeType, durationType:
		return true
	}
	return false
}

// These require special functions for encoding/decoding.
func isJSONWellKnownType(rt reflect.Type) (wellKnown bool) {
	// Special cases based on type.
	switch rt {
	// Native types.
	case timeType, durationType:
		return true
	// Google "well known" types.
	case
		gAnyType, gTimestampType, gDurationType, gEmptyType,
		/*gStructType, gValueType, gListType,*/
		/*gDoubleType, gFloatType,*/
		gInt64Type, gUInt64Type, gInt32Type, gUInt32Type, gBoolType,
		gStringType, gBytesType:
		return true
	}
	// General cases based on kind.
	switch rt.Kind() {
	case
		reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		reflect.Array, reflect.Slice, reflect.String:
		return true
	default:
		return false
	}
}

// Returns ok=false if nothing was done because the default behavior is fine (or if err).
// TODO: remove proto dependency.
func encodeReflectJSONWellKnown(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (ok bool, err error) {
	switch info.Type {
	// Native types.
	case timeType:
		// See https://github.com/golang/protobuf/blob/d04d7b157bb510b1e0c10132224b616ac0e26b17/jsonpb/encode.go#L308,
		// "RFC 3339, where generated output will always be Z-normalized
		//  and uses 0, 3, 6 or 9 fractional digits."
		t := rv.Interface().(time.Time)
		err = EncodeJSONTime(w, t)
		if err != nil {
			return false, err
		}
		return true, nil
	case durationType:
		// "Generated output always contains 0, 3, 6, or 9 fractional digits,
		//  depending on required precision."
		d := rv.Interface().(time.Duration)
		err = EncodeJSONDuration(w, d)
		if err != nil {
			return false, err
		}
		return true, nil
	// Google "well known" types.
	// The protobuf Timestamp and Duration values contain a Mutex, and therefore must not be copied.
	// The corresponding reflect value may not be addressable, we can not safely get their pointer.
	// So we just extract the `Seconds` and `Nanos` fields from the reflect value, without copying
	// the whole struct, and encode them as their coresponding time.Time or time.Duration value.
	case gTimestampType:
		t := time.Unix(rv.Interface().(timestamppb.Timestamp).Seconds, int64(rv.Interface().(timestamppb.Timestamp).Nanos))
		err = EncodeJSONTime(w, t)
		if err != nil {
			return false, err
		}
		return true, nil
	case gDurationType:
		d := time.Duration(rv.Interface().(durationpb.Duration).Seconds) * time.Second
		d += time.Duration(rv.Interface().(durationpb.Duration).Nanos)
		err = EncodeJSONDuration(w, d)
		if err != nil {
			return false, err
		}
		return true, nil
	// TODO: port each below to above without proto dependency
	// for marshaling code, to minimize dependencies.
	case
		gAnyType, gEmptyType,
		/*gStructType, gValueType, gListType,*/
		/*gDoubleType, gFloatType,*/
		gInt64Type, gUInt64Type, gInt32Type, gUInt32Type, gBoolType,
		gStringType, gBytesType:
		bz, err := proto.Marshal(rv.Interface().(proto.Message))
		if err != nil {
			return false, err
		}
		_, err = w.Write(bz)
		return true, err
	}
	return false, nil
}

// Returns ok=false if nothing was done because the default behavior is fine.
// CONTRACT: rv is a concrete type.
func decodeReflectJSONWellKnown(bz []byte, info *TypeInfo, rv reflect.Value, fopts FieldOptions) (ok bool, err error) {
	if rv.Kind() == reflect.Interface {
		panic("expected a concrete type to decode to")
	}
	switch info.Type {
	// Native types.
	case timeType:
		var t time.Time
		t, err = DecodeJSONTime(bz, fopts)
		if err != nil {
			return false, err
		}
		rv.Set(reflect.ValueOf(t))
		return true, nil
	case durationType:
		var d time.Duration
		d, err = DecodeJSONDuration(bz, fopts)
		if err != nil {
			return false, err
		}
		rv.Set(reflect.ValueOf(d))
		return true, nil
	// Google "well known" types.
	case gTimestampType:
		var t timestamppb.Timestamp
		t, err = DecodeJSONPBTimestamp(bz, fopts)
		if err != nil {
			return false, err
		}
		rv.FieldByName("Seconds").Set(reflect.ValueOf(t.Seconds))
		rv.FieldByName("Nanos").Set(reflect.ValueOf(t.Nanos))
		return true, nil
	case gDurationType:
		var d durationpb.Duration
		d, err = DecodeJSONPBDuration(bz, fopts)
		if err != nil {
			return false, err
		}
		rv.FieldByName("Seconds").Set(reflect.ValueOf(d.Seconds))
		rv.FieldByName("Nanos").Set(reflect.ValueOf(d.Nanos))
		return true, nil
	// TODO: port each below to above without proto dependency
	// for unmarshaling code, to minimize dependencies.
	case
		gAnyType, gEmptyType,
		/*gStructType, gValueType, gListType,*/
		/*gDoubleType, gFloatType,*/
		gInt64Type, gUInt64Type, gInt32Type, gUInt32Type, gBoolType,
		gStringType, gBytesType:
		err := proto.Unmarshal(bz, rv.Addr().Interface().(proto.Message))
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Returns ok=false if nothing was done because the default behavior is fine.
func encodeReflectBinaryWellKnown(w io.Writer, info *TypeInfo, rv reflect.Value, fopts FieldOptions, bare bool) (ok bool, err error) {
	// Validations.
	if rv.Kind() == reflect.Interface {
		panic("expected a concrete type to decode to")
	}
	// Maybe recurse with length-prefixing.
	if !bare {
		buf := poolBytesBuffer.Get()
		defer poolBytesBuffer.Put(buf)

		ok, err = encodeReflectBinaryWellKnown(buf, info, rv, fopts, true)
		if err != nil {
			return false, err
		}
		err = EncodeByteSlice(w, buf.Bytes())
		if err != nil {
			return false, err
		}
		return true, nil
	}
	switch info.Type {
	// Native types.
	case timeType:
		var t time.Time
		t = rv.Interface().(time.Time)
		err = EncodeTime(w, t)
		if err != nil {
			return false, err
		}
		return true, nil
	case durationType:
		var d time.Duration
		d = rv.Interface().(time.Duration)
		err = EncodeDuration(w, d)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Returns ok=false if nothing was done because the default behavior is fine.
func decodeReflectBinaryWellKnown(bz []byte, info *TypeInfo, rv reflect.Value, fopts FieldOptions, bare bool) (ok bool, n int, err error) {
	// Validations.
	if rv.Kind() == reflect.Interface {
		panic("expected a concrete type to decode to")
	}
	// Strip if needed.
	bz, err = decodeMaybeBare(bz, &n, bare)
	if err != nil {
		return false, n, err
	}
	switch info.Type {
	// Native types.
	case timeType:
		var t time.Time
		var n_ int
		t, n_, err = DecodeTime(bz)
		if slide(&bz, &n, n_) && err != nil {
			return false, n, err
		}
		rv.Set(reflect.ValueOf(t))
		return true, n, nil
	case durationType:
		var d time.Duration
		var n_ int
		d, n_, err = DecodeDuration(bz)
		if slide(&bz, &n, n_) && err != nil {
			return false, n, err
		}
		rv.Set(reflect.ValueOf(d))
		return true, n, nil
	}
	return false, 0, nil
}

//----------------------------------------
// Well known JSON encoders and decoders

func EncodeJSONTimeValue(w io.Writer, s int64, ns int32) (err error) {
	err = validateTimeValue(s, ns)
	if err != nil {
		return err
	}
	// time.RFC3339Nano isn't exactly right (we need to get 3/6/9 fractional digits).
	t := time.Unix(s, int64(ns)).Round(0).UTC()
	x := t.Format("2006-01-02T15:04:05.000000000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, ".000")
	_, err = fmt.Fprintf(w, `"%vZ"`, x)
	return err
}

func EncodeJSONTime(w io.Writer, t time.Time) (err error) {
	t = t.Round(0).UTC()
	return EncodeJSONTimeValue(w, t.Unix(), int32(t.Nanosecond()))
}

func EncodeJSONPBTimestamp(w io.Writer, t *timestamppb.Timestamp) (err error) {
	return EncodeJSONTimeValue(w, t.GetSeconds(), t.GetNanos())
}

func EncodeJSONDurationValue(w io.Writer, s int64, ns int32) (err error) {
	err = validateDurationValue(s, ns)
	if err != nil {
		return err
	}
	sign := ""
	if s < 0 {
		s = -s
		sign = "-"
	}
	if ns < 0 {
		ns = -ns
		sign = "-" // could be true even if s == 0.
	}
	x := fmt.Sprintf("%s%d.%09d", sign, s, ns)
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, ".000")
	_, err = fmt.Fprintf(w, `"%vs"`, x)
	return err
}

func EncodeJSONDuration(w io.Writer, d time.Duration) (err error) {
	return EncodeJSONDurationValue(w, int64(d)/1e9, int32(int64(d)%1e9))
}

func EncodeJSONPBDuration(w io.Writer, d *durationpb.Duration) (err error) {
	return EncodeJSONDurationValue(w, d.GetSeconds(), d.GetNanos())
}

func DecodeJSONTime(bz []byte, fopts FieldOptions) (t time.Time, err error) {
	t = emptyTime // defensive
	v, err := unquoteString(string(bz))
	if err != nil {
		return
	}
	t, err = time.Parse(time.RFC3339Nano, v)
	if err != nil {
		err = fmt.Errorf("bad time: %w", err)
		return
	}
	return
}

// NOTE: probably not needed after protobuf v1.25 and after, replace with New().
func newPBTimestamp(t time.Time) timestamppb.Timestamp {
	return timestamppb.Timestamp{Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}
}

func DecodeJSONPBTimestamp(bz []byte, fopts FieldOptions) (t timestamppb.Timestamp, err error) {
	var t_ time.Time
	t_, err = DecodeJSONTime(bz, fopts)
	if err != nil {
		return
	}
	return newPBTimestamp(t_), nil
}

func DecodeJSONDuration(bz []byte, fopts FieldOptions) (d time.Duration, err error) {
	v, err := unquoteString(string(bz))
	if err != nil {
		return
	}
	d, err = time.ParseDuration(v)
	if err != nil {
		err = fmt.Errorf("bad time: %w", err)
		return
	}
	return
}

// NOTE: probably not needed after protobuf v1.25 and after, replace with New().
func newPBDuration(d time.Duration) durationpb.Duration {
	nanos := d.Nanoseconds()
	secs := nanos / 1e9
	nanos -= secs * 1e9
	return durationpb.Duration{Seconds: secs, Nanos: int32(nanos)}
}

func DecodeJSONPBDuration(bz []byte, fopts FieldOptions) (d durationpb.Duration, err error) {
	var d_ time.Duration
	d_, err = DecodeJSONDuration(bz, fopts)
	if err != nil {
		return
	}
	return newPBDuration(d_), nil
}

func IsEmptyTime(t time.Time) bool {
	t = t.Round(0).UTC()
	return t.Unix() == 0 && t.Nanosecond() == 0
}

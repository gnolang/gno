package amino

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/bits"
	"time"
)

// ----------------------------------------
// Signed

func EncodeVarint8(w io.Writer, i int8) (err error) {
	var buf [2]byte
	n := binary.PutVarint(buf[:], int64(i))
	_, err = w.Write(buf[0:n])
	return
}

func EncodeVarint16(w io.Writer, i int16) (err error) {
	var buf [3]byte
	n := binary.PutVarint(buf[:], int64(i))
	_, err = w.Write(buf[0:n])
	return
}

func EncodeVarint32(w io.Writer, i int32) (err error) {
	var buf [5]byte
	n := binary.PutVarint(buf[:], int64(i))
	_, err = w.Write(buf[0:n])
	return
}

func EncodeVarint(w io.Writer, i int64) (err error) {
	var buf [10]byte
	n := binary.PutVarint(buf[:], i)
	_, err = w.Write(buf[0:n])
	return
}

func EncodeInt32(w io.Writer, i int32) (err error) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(i))
	_, err = w.Write(buf[:])
	return
}

func EncodeInt64(w io.Writer, i int64) (err error) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(i))
	_, err = w.Write(buf[:])
	return err
}

func VarintSize(i int64) int {
	return UvarintSize((uint64(i) << 1) ^ uint64(i>>63))
}

// ----------------------------------------
// Unsigned

// Unlike EncodeUint8, writes a single byte.
func EncodeByte(w io.Writer, b byte) (err error) {
	_, err = w.Write([]byte{b})
	return
}

func EncodeUvarint8(w io.Writer, i uint8) (err error) {
	var buf [2]byte
	n := binary.PutUvarint(buf[:], uint64(i))
	_, err = w.Write(buf[0:n])
	return
}

func EncodeUvarint16(w io.Writer, i uint16) (err error) {
	var buf [3]byte
	n := binary.PutUvarint(buf[:], uint64(i))
	_, err = w.Write(buf[0:n])
	return
}

func EncodeUvarint32(w io.Writer, i uint32) (err error) {
	var buf [5]byte
	n := binary.PutUvarint(buf[:], uint64(i))
	_, err = w.Write(buf[0:n])
	return
}

func EncodeUvarint(w io.Writer, u uint64) (err error) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], u)
	_, err = w.Write(buf[0:n])
	return
}

func EncodeUint32(w io.Writer, u uint32) (err error) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], u)
	_, err = w.Write(buf[:])
	return
}

func EncodeUint64(w io.Writer, u uint64) (err error) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], u)
	_, err = w.Write(buf[:])
	return
}

func UvarintSize(u uint64) int {
	if u == 0 {
		return 1
	}
	return (bits.Len64(u) + 6) / 7
}

// ----------------------------------------
// Other Primitives

func EncodeBool(w io.Writer, b bool) (err error) {
	if b {
		err = EncodeByte(w, 0x01)
	} else {
		err = EncodeByte(w, 0x00)
	}
	return
}

// NOTE: UNSAFE
func EncodeFloat32(w io.Writer, f float32) (err error) {
	return EncodeUint32(w, math.Float32bits(f))
}

// NOTE: UNSAFE
func EncodeFloat64(w io.Writer, f float64) (err error) {
	return EncodeUint64(w, math.Float64bits(f))
}

// ----------------------------------------
// Time and Duration

const (
	// See https://github.com/protocolbuffers/protobuf/blob/d2980062c859649523d5fd51d6b55ab310e47482/src/google/protobuf/timestamp.proto#L123-L135
	// seconds of 01-01-0001
	minTimeSeconds int64 = -62135596800
	// seconds of 10000-01-01
	maxTimeSeconds int64 = 253402300800 // exclusive
	// nanos have to be in interval: [0, 999999999]
	maxTimeNanos = 999999999 // inclusive

	// See https://github.com/protocolbuffers/protobuf/blob/d2980062c859649523d5fd51d6b55ab310e47482/src/google/protobuf/duration.proto#L105-L116
	minDurationSeconds int64 = -315576000000
	maxDurationSeconds int64 = 315576000000 // inclusive
	minDurationNanos         = -999999999
	maxDurationNanos         = 999999999 // inclusive
)

type InvalidTimeError string

func (e InvalidTimeError) Error() string {
	return "invalid time: " + string(e)
}

type InvalidDurationError string

func (e InvalidDurationError) Error() string {
	return "invalid duration: " + string(e)
}

// EncodeTimeValue writes the number of seconds (int64) and nanoseconds (int32),
// with millisecond resolution since January 1, 1970 UTC to the Writer as an
// UInt64.
// Milliseconds are used to ease compatibility with Javascript,
// which does not support finer resolution.
/* See https://godoc.org/google.golang.org/protobuf/types/known/timestamppb#Timestamp
type Timestamp struct {

    // Represents seconds of UTC time since Unix epoch
    // 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to
    // 9999-12-31T23:59:59Z inclusive.
    Seconds int64 `protobuf:"varint,1,opt,name=seconds,proto3" json:"seconds,omitempty"`
    // Non-negative fractions of a second at nanosecond resolution. Negative
    // second values with fractions must still have non-negative nanos values
    // that count forward in time. Must be from 0 to 999,999,999
    // inclusive.
    Nanos int32 `protobuf:"varint,2,opt,name=nanos,proto3" json:"nanos,omitempty"`
    // contains filtered or unexported fields
}
*/
func EncodeTimeValue(w io.Writer, s int64, ns int32) (err error) {
	// Validations
	err = validateTimeValue(s, ns)
	if err != nil {
		return
	}
	// skip if default/zero value:
	if s != 0 {
		err = encodeFieldNumberAndTyp3(w, 1, Typ3Varint)
		if err != nil {
			return
		}
		err = EncodeUvarint(w, uint64(s))
		if err != nil {
			return
		}
	}
	// skip if default/zero value:
	if ns != 0 {
		err = encodeFieldNumberAndTyp3(w, 2, Typ3Varint)
		if err != nil {
			return
		}
		err = EncodeUvarint(w, uint64(ns))
		if err != nil {
			return
		}
	}

	return err
}

func EncodeTime(w io.Writer, t time.Time) (err error) {
	return EncodeTimeValue(w, t.Unix(), int32(t.Nanosecond()))
}

func validateTimeValue(s int64, ns int32) (err error) {
	if s < minTimeSeconds || s >= maxTimeSeconds {
		return InvalidTimeError(fmt.Sprintf("seconds have to be >= %d and < %d, got: %d",
			minTimeSeconds, maxTimeSeconds, s))
	}
	if ns < 0 || ns > maxTimeNanos {
		// we could as well panic here:
		// time.Time.Nanosecond() guarantees nanos to be in [0, 999,999,999]
		return InvalidTimeError(fmt.Sprintf("nanoseconds have to be >= 0 and <= %v, got: %d",
			maxTimeNanos, ns))
	}
	return nil
}

// The binary encoding of Duration is the same as Timestamp,
// but the validation checks are different.
/* See https://godoc.org/google.golang.org/protobuf/types/known/durationpb#Duration
type Duration struct {

    // Signed seconds of the span of time. Must be from -315,576,000,000
    // to +315,576,000,000 inclusive. Note: these bounds are computed from:
    // 60 sec/min * 60 min/hr * 24 hr/day * 365.25 days/year * 10000 years
    Seconds int64 `protobuf:"varint,1,opt,name=seconds,proto3" json:"seconds,omitempty"`
    // Signed fractions of a second at nanosecond resolution of the span
    // of time. Durations less than one second are represented with a 0
    // `seconds` field and a positive or negative `nanos` field. For durations
    // of one second or more, a non-zero value for the `nanos` field must be
    // of the same sign as the `seconds` field. Must be from -999,999,999
    // to +999,999,999 inclusive.
    Nanos int32 `protobuf:"varint,2,opt,name=nanos,proto3" json:"nanos,omitempty"`
    // contains filtered or unexported fields
}
*/
func EncodeDurationValue(w io.Writer, s int64, ns int32) (err error) {
	// Validations
	err = validateDurationValue(s, ns)
	if err != nil {
		return err
	}
	// skip if default/zero value:
	if s != 0 {
		err = encodeFieldNumberAndTyp3(w, 1, Typ3Varint)
		if err != nil {
			return
		}
		err = EncodeUvarint(w, uint64(s))
		if err != nil {
			return
		}
	}
	// skip if default/zero value:
	if ns != 0 {
		err = encodeFieldNumberAndTyp3(w, 2, Typ3Varint)
		if err != nil {
			return
		}
		err = EncodeUvarint(w, uint64(ns))
		if err != nil {
			return
		}
	}

	return err
}

func EncodeDuration(w io.Writer, d time.Duration) (err error) {
	sns := d.Nanoseconds()
	s, ns := sns/1e9, int32(sns%1e9)
	err = validateDurationValue(s, ns)
	if err != nil {
		return err
	}
	return EncodeDurationValue(w, s, ns)
}

func validateDurationValue(s int64, ns int32) (err error) {
	if (s > 0 && ns < 0) || (s < 0 && ns > 0) {
		return InvalidDurationError(fmt.Sprintf("signs of seconds and nanos do not match: %v and %v",
			s, ns))
	}
	if s < minDurationSeconds || s > maxDurationSeconds {
		return InvalidDurationError(fmt.Sprintf("seconds have to be >= %d and < %d, got: %d",
			minDurationSeconds, maxDurationSeconds, s))
	}
	if ns < minDurationNanos || ns > maxDurationNanos {
		return InvalidDurationError(fmt.Sprintf("ns out of range [%v, %v], got: %v",
			minDurationNanos, maxDurationNanos, ns))
	}
	return nil
}

const (
	// On the other hand, Go's native duration only allows a smaller interval:
	// https://golang.org/pkg/time/#Duration
	minDurationSecondsGo = int64(math.MinInt64) / int64(1e9)
	maxDurationSecondsGo = int64(math.MaxInt64) / int64(1e9)
)

// Go's time.Duration has a more limited range.
// This is specific to Go and not Amino.
func validateDurationValueGo(s int64, ns int32) (err error) {
	err = validateDurationValue(s, ns)
	if err != nil {
		return err
	}
	if s < minDurationSecondsGo || s > maxDurationSecondsGo {
		return InvalidDurationError(fmt.Sprintf("duration seconds exceeds bounds for Go's time.Duration type: %v",
			s))
	}
	sns := s*1e9 + int64(ns)
	if sns > 0 && s < 0 || sns < 0 && s > 0 {
		return InvalidDurationError(fmt.Sprintf("duration seconds+nanoseconds exceeds bounds for Go's time.Duration type: %v and %v",
			s, ns))
	}
	return nil
}

// ----------------------------------------
// Byte Slices and Strings

func EncodeByteSlice(w io.Writer, bz []byte) (err error) {
	err = EncodeUvarint(w, uint64(len(bz)))
	if err != nil {
		return
	}
	_, err = w.Write(bz)
	return
}

func ByteSliceSize(bz []byte) int {
	return UvarintSize(uint64(len(bz))) + len(bz)
}

func EncodeString(w io.Writer, s string) (err error) {
	return EncodeByteSlice(w, []byte(s))
}

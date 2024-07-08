package amino

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

// ----------------------------------------
// Signed

func DecodeVarint8(bz []byte) (i int8, n int, err error) {
	i64, n, err := DecodeVarint(bz)
	if err != nil {
		return
	}
	if i64 < int64(math.MinInt8) || i64 > int64(math.MaxInt8) {
		err = errors.New("EOF decoding int8")
		return
	}
	i = int8(i64)
	return
}

func DecodeVarint16(bz []byte) (i int16, n int, err error) {
	i64, n, err := DecodeVarint(bz)
	if err != nil {
		return
	}
	if i64 < int64(math.MinInt16) || i64 > int64(math.MaxInt16) {
		err = errors.New("EOF decoding int16")
		return
	}
	i = int16(i64)
	return
}

func DecodeVarint(bz []byte) (i int64, n int, err error) {
	i, n = binary.Varint(bz)
	if n == 0 {
		// buf too small
		err = errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		err = errors.New("EOF decoding varint")
	}
	return
}

func DecodeInt32(bz []byte) (i int32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		err = errors.New("EOF decoding int32")
		return
	}
	i = int32(binary.LittleEndian.Uint32(bz[:size]))
	n = size
	return
}

func DecodeInt64(bz []byte) (i int64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		err = errors.New("EOF decoding int64")
		return
	}
	i = int64(binary.LittleEndian.Uint64(bz[:size]))
	n = size
	return
}

// ----------------------------------------
// Unsigned

func DecodeByte(bz []byte) (b byte, n int, err error) {
	if len(bz) == 0 {
		err = errors.New("EOF decoding byte")
		return
	}
	b = bz[0]
	n = 1
	return
}

func DecodeUvarint8(bz []byte) (u uint8, n int, err error) {
	u64, n, err := DecodeUvarint(bz)
	if err != nil {
		return
	}
	if u64 > uint64(math.MaxUint8) {
		err = errors.New("EOF decoding uint8")
		return
	}
	u = uint8(u64)
	return
}

func DecodeUvarint16(bz []byte) (u uint16, n int, err error) {
	u64, n, err := DecodeUvarint(bz)
	if err != nil {
		return
	}
	if u64 > uint64(math.MaxUint16) {
		err = errors.New("EOF decoding uint16")
		return
	}
	u = uint16(u64)
	return
}

func DecodeUvarint32(bz []byte) (u uint32, n int, err error) {
	u64, n, err := DecodeUvarint(bz)
	if err != nil {
		return
	}
	if u64 > uint64(math.MaxUint32) {
		err = errors.New("EOF decoding uint32")
		return
	}
	u = uint32(u64)
	return
}

func DecodeUvarint(bz []byte) (u uint64, n int, err error) {
	u, n = binary.Uvarint(bz)
	if n == 0 {
		// buf too small
		err = errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		err = errors.New("EOF decoding uvarint")
	}
	return
}

func DecodeUint32(bz []byte) (u uint32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		err = errors.New("EOF decoding uint32")
		return
	}
	u = binary.LittleEndian.Uint32(bz[:size])
	n = size
	return
}

func DecodeUint64(bz []byte) (u uint64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		err = errors.New("EOF decoding uint64")
		return
	}
	u = binary.LittleEndian.Uint64(bz[:size])
	n = size
	return
}

// ----------------------------------------
// Other Primitives

func DecodeBool(bz []byte) (b bool, n int, err error) {
	const size int = 1
	if len(bz) < size {
		err = errors.New("EOF decoding bool")
		return
	}
	switch bz[0] {
	case 0:
		b = false
	case 1:
		b = true
	default:
		err = errors.New("invalid bool")
	}
	n = size
	return
}

// NOTE: UNSAFE
func DecodeFloat32(bz []byte) (f float32, n int, err error) {
	const size int = 4
	if len(bz) < size {
		err = errors.New("EOF decoding float32")
		return
	}
	i := binary.LittleEndian.Uint32(bz[:size])
	f = math.Float32frombits(i)
	n = size
	return
}

// NOTE: UNSAFE
func DecodeFloat64(bz []byte) (f float64, n int, err error) {
	const size int = 8
	if len(bz) < size {
		err = errors.New("EOF decoding float64")
		return
	}
	i := binary.LittleEndian.Uint64(bz[:size])
	f = math.Float64frombits(i)
	n = size
	return
}

// ----------------------------------------
// Time and Duration

// DecodeTimeValue decodes seconds (int64) and nanoseconds (int32) since January 1,
// 1970 UTC, and returns the corresponding time.  If nanoseconds is not in the
// range [0, 999999999], or if seconds is too large, an error is returned.
func DecodeTimeValue(bz []byte) (s int64, ns int32, n int, err error) {
	// Read sec and nanosec.
	s, n, err = decodeSeconds(&bz)
	if err != nil {
		return
	}
	ns, err = decodeNanos(&bz, &n)
	if err != nil {
		return
	}
	// Validations
	err = validateTimeValue(s, ns)
	if err != nil {
		return
	}
	return
}

func DecodeTime(bz []byte) (t time.Time, n int, err error) {
	// Defensively set default to to emptyTime (1970, not 0001)
	t = emptyTime
	s, ns, n, err := DecodeTimeValue(bz)
	if err != nil {
		return
	}
	// Construct time.
	t = time.Unix(s, int64(ns))
	// Strip timezone and monotonic for deep equality.
	t = t.UTC().Truncate(0)
	return
}

func DecodeDurationValue(bz []byte) (s int64, ns int32, n int, err error) {
	// Read sec and nanosec.
	s, n, err = decodeSeconds(&bz)
	if err != nil {
		return
	}
	ns, err = decodeNanos(&bz, &n)
	if err != nil {
		return
	}
	// Validations
	err = validateDurationValue(s, ns)
	if err != nil {
		return
	}
	return
}

func DecodeDuration(bz []byte) (d time.Duration, n int, err error) {
	// Defensively set default to to zeroDuration
	s, ns, n, err := DecodeDurationValue(bz)
	if err != nil {
		return
	}
	// Validations
	err = validateDurationValueGo(s, ns)
	if err != nil {
		return
	}
	// Construct Duration.
	d = time.Duration(s*1e9 + int64(ns))
	return
}

// If bz is empty, returns err=nil.
// Does not validate.
func decodeSeconds(bz *[]byte) (int64, int, error) {
	if len(*bz) == 0 {
		return 0, 0, nil
	}
	// Optionally decode field number 1 and Typ3 (8Byte).
	// only slide if we need to:
	var n int
	fieldNum, typ, _n, err := decodeFieldNumberAndTyp3(*bz)
	if err != nil {
		return 0, n, err
	}
	switch {
	case fieldNum == 1 && typ == Typ3Varint:
		slide(bz, &n, _n)
		sec, _n, err := DecodeUvarint(*bz)
		if slide(bz, &n, _n) && err != nil {
			return 0, n, err
		}
		// if seconds where negative before casting them to uint64, we yield
		// the original signed value:
		res := int64(sec)
		return res, n, err
	case fieldNum == 2 && typ == Typ3Varint:
		// skip: do not slide, no error, will read again
		return 0, n, nil
	default:
		return 0, n, fmt.Errorf("expected field number 1 <Varint> or field number 2 <Varint> , got %v", fieldNum)
	}
}

// If bz is empty, returns err=nil.
// Validates whether ns is in range, but caller may want to check for non-negativity.
func decodeNanos(bz *[]byte, n *int) (int32, error) {
	if len(*bz) == 0 {
		return 0, nil
	}
	// Optionally decode field number 2 and Typ3 (4Byte).
	fieldNum, typ, _n, err := decodeFieldNumberAndTyp3(*bz)
	if err != nil {
		return 0, err
	}
	if fieldNum == 2 && typ == Typ3Varint {
		slide(bz, n, _n)
		nsec_, _n, err := DecodeUvarint(*bz)
		if slide(bz, n, _n) && err != nil {
			return 0, err
		}
		nsec := int64(nsec_)
		// Validation check.
		if 1e9 <= nsec || nsec <= -1e9 {
			return 0, InvalidTimeError(fmt.Sprintf("nanoseconds not in interval [-999999999, 999999999] %v", nsec))
		}
		// this cast from uint64 to int32 is OK, due to above restriction:
		return int32(nsec), nil
	}
	// skip over (no error)
	return 0, nil
}

// ----------------------------------------
// Byte Slices and Strings

func DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	var count uint64
	var _n int
	count, _n, err = DecodeUvarint(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	if int(count) < 0 {
		err = fmt.Errorf("invalid negative length %v decoding []byte", count)
		return
	}
	if len(bz) < int(count) {
		err = fmt.Errorf("insufficient bytes decoding []byte of length %v: %X", count, bz)
		return
	}
	bz2 = make([]byte, count)
	copy(bz2, bz[0:count])
	n += int(count)
	return
}

func DecodeString(bz []byte) (s string, n int, err error) {
	var bz2 []byte
	bz2, n, err = DecodeByteSlice(bz)
	s = string(bz2)
	return
}

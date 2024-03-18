package uint256

import (
	"encoding/binary"
)

func (z *Uint) SetBytes1(in []byte) *Uint {
	z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
	z.arr[0] = uint64(in[0])
	return z
}

func (z *Uint) SetBytes2(in []byte) *Uint {
	_ = in[1] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
	z.arr[0] = uint64(binary.BigEndian.Uint16(in[0:2]))
	return z
}

func (z *Uint) SetBytes3(in []byte) *Uint {
	_ = in[2] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
	z.arr[0] = uint64(binary.BigEndian.Uint16(in[1:3])) | uint64(in[0])<<16
	return z
}

func (z *Uint) SetBytes4(in []byte) *Uint {
	_ = in[3] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
	z.arr[0] = uint64(binary.BigEndian.Uint32(in[0:4]))
	return z
}

func (z *Uint) SetBytes5(in []byte) *Uint {
	_ = in[4] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
	z.arr[0] = bigEndianUint40(in[0:5])
	return z
}

func (z *Uint) SetBytes6(in []byte) *Uint {
	_ = in[5] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
	z.arr[0] = bigEndianUint48(in[0:6])
	return z
}

func (z *Uint) SetBytes7(in []byte) *Uint {
	_ = in[6] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
	z.arr[0] = bigEndianUint56(in[0:7])
	return z
}

// SetBytes8 is identical to SetBytes(in[:8]), but panics is input is too short
func (z *Uint) SetBytes8(in []byte) *Uint {
	_ = in[7] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
	z.arr[0] = binary.BigEndian.Uint64(in[0:8])
	return z
}

func (z *Uint) SetBytes9(in []byte) *Uint {
	_ = in[8] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2] = 0, 0
	z.arr[1] = uint64(in[0])
	z.arr[0] = binary.BigEndian.Uint64(in[1:9])
	return z
}

func (z *Uint) SetBytes10(in []byte) *Uint {
	_ = in[9] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2] = 0, 0
	z.arr[1] = uint64(binary.BigEndian.Uint16(in[0:2]))
	z.arr[0] = binary.BigEndian.Uint64(in[2:10])
	return z
}

func (z *Uint) SetBytes11(in []byte) *Uint {
	_ = in[10] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2] = 0, 0
	z.arr[1] = uint64(binary.BigEndian.Uint16(in[1:3])) | uint64(in[0])<<16
	z.arr[0] = binary.BigEndian.Uint64(in[3:11])
	return z
}


func (z *Uint) SetBytes12(in []byte) *Uint {
	_ = in[11] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2] = 0, 0
	z.arr[1] = uint64(binary.BigEndian.Uint32(in[0:4]))
	z.arr[0] = binary.BigEndian.Uint64(in[4:12])
	return z
}

func (z *Uint) SetBytes13(in []byte) *Uint {
	_ = in[12] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2] = 0, 0
	z.arr[1] = bigEndianUint40(in[0:5])
	z.arr[0] = binary.BigEndian.Uint64(in[5:13])
	return z
}

func (z *Uint) SetBytes14(in []byte) *Uint {
	_ = in[13] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2] = 0, 0
	z.arr[1] = bigEndianUint48(in[0:6])
	z.arr[0] = binary.BigEndian.Uint64(in[6:14])
	return z
}

func (z *Uint) SetBytes15(in []byte) *Uint {
	_ = in[14] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2] = 0, 0
	z.arr[1] = bigEndianUint56(in[0:7])
	z.arr[0] = binary.BigEndian.Uint64(in[7:15])
	return z
}

// SetBytes16 is identical to SetBytes(in[:16]), but panics is input is too short
func (z *Uint) SetBytes16(in []byte) *Uint {
	_ = in[15] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3], z.arr[2] = 0, 0
	z.arr[1] = binary.BigEndian.Uint64(in[0:8])
	z.arr[0] = binary.BigEndian.Uint64(in[8:16])
	return z
}

func (z *Uint) SetBytes17(in []byte) *Uint {
	_ = in[16] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = 0
	z.arr[2] = uint64(in[0])
	z.arr[1] = binary.BigEndian.Uint64(in[1:9])
	z.arr[0] = binary.BigEndian.Uint64(in[9:17])
	return z
}

func (z *Uint) SetBytes18(in []byte) *Uint {
	_ = in[17] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = 0
	z.arr[2] = uint64(binary.BigEndian.Uint16(in[0:2]))
	z.arr[1] = binary.BigEndian.Uint64(in[2:10])
	z.arr[0] = binary.BigEndian.Uint64(in[10:18])
	return z
}

func (z *Uint) SetBytes19(in []byte) *Uint {
	_ = in[18] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = 0
	z.arr[2] = uint64(binary.BigEndian.Uint16(in[1:3])) | uint64(in[0])<<16
	z.arr[1] = binary.BigEndian.Uint64(in[3:11])
	z.arr[0] = binary.BigEndian.Uint64(in[11:19])
	return z
}

func (z *Uint) SetBytes20(in []byte) *Uint {
	_ = in[19] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = 0
	z.arr[2] = uint64(binary.BigEndian.Uint32(in[0:4]))
	z.arr[1] = binary.BigEndian.Uint64(in[4:12])
	z.arr[0] = binary.BigEndian.Uint64(in[12:20])
	return z
}

func (z *Uint) SetBytes21(in []byte) *Uint {
	_ = in[20] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = 0
	z.arr[2] = bigEndianUint40(in[0:5])
	z.arr[1] = binary.BigEndian.Uint64(in[5:13])
	z.arr[0] = binary.BigEndian.Uint64(in[13:21])
	return z
}

func (z *Uint) SetBytes22(in []byte) *Uint {
	_ = in[21] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = 0
	z.arr[2] = bigEndianUint48(in[0:6])
	z.arr[1] = binary.BigEndian.Uint64(in[6:14])
	z.arr[0] = binary.BigEndian.Uint64(in[14:22])
	return z
}

func (z *Uint) SetBytes23(in []byte) *Uint {
	_ = in[22] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = 0
	z.arr[2] = bigEndianUint56(in[0:7])
	z.arr[1] = binary.BigEndian.Uint64(in[7:15])
	z.arr[0] = binary.BigEndian.Uint64(in[15:23])
	return z
}

// SetBytes24 is identical to SetBytes(in[:24]), but panics is input is too short
func (z *Uint) SetBytes24(in []byte) *Uint {
	_ = in[23] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = 0
	z.arr[2] = binary.BigEndian.Uint64(in[0:8])
	z.arr[1] = binary.BigEndian.Uint64(in[8:16])
	z.arr[0] = binary.BigEndian.Uint64(in[16:24])
	return z
}

func (z *Uint) SetBytes25(in []byte) *Uint {
	_ = in[24] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = uint64(in[0])
	z.arr[2] = binary.BigEndian.Uint64(in[1:9])
	z.arr[1] = binary.BigEndian.Uint64(in[9:17])
	z.arr[0] = binary.BigEndian.Uint64(in[17:25])
	return z
}

func (z *Uint) SetBytes26(in []byte) *Uint {
	_ = in[25] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = uint64(binary.BigEndian.Uint16(in[0:2]))
	z.arr[2] = binary.BigEndian.Uint64(in[2:10])
	z.arr[1] = binary.BigEndian.Uint64(in[10:18])
	z.arr[0] = binary.BigEndian.Uint64(in[18:26])
	return z
}

func (z *Uint) SetBytes27(in []byte) *Uint {
	_ = in[26] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = uint64(binary.BigEndian.Uint16(in[1:3])) | uint64(in[0])<<16
	z.arr[2] = binary.BigEndian.Uint64(in[3:11])
	z.arr[1] = binary.BigEndian.Uint64(in[11:19])
	z.arr[0] = binary.BigEndian.Uint64(in[19:27])
	return z
}

func (z *Uint) SetBytes28(in []byte) *Uint {
	_ = in[27] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = uint64(binary.BigEndian.Uint32(in[0:4]))
	z.arr[2] = binary.BigEndian.Uint64(in[4:12])
	z.arr[1] = binary.BigEndian.Uint64(in[12:20])
	z.arr[0] = binary.BigEndian.Uint64(in[20:28])
	return z
}

func (z *Uint) SetBytes29(in []byte) *Uint {
	_ = in[23] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = bigEndianUint40(in[0:5])
	z.arr[2] = binary.BigEndian.Uint64(in[5:13])
	z.arr[1] = binary.BigEndian.Uint64(in[13:21])
	z.arr[0] = binary.BigEndian.Uint64(in[21:29])
	return z
}


func (z *Uint) SetBytes30(in []byte) *Uint {
	_ = in[29] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = bigEndianUint48(in[0:6])
	z.arr[2] = binary.BigEndian.Uint64(in[6:14])
	z.arr[1] = binary.BigEndian.Uint64(in[14:22])
	z.arr[0] = binary.BigEndian.Uint64(in[22:30])
	return z
}

func (z *Uint) SetBytes31(in []byte) *Uint {
	_ = in[30] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = bigEndianUint56(in[0:7])
	z.arr[2] = binary.BigEndian.Uint64(in[7:15])
	z.arr[1] = binary.BigEndian.Uint64(in[15:23])
	z.arr[0] = binary.BigEndian.Uint64(in[23:31])
	return z
}

func (z *Uint) SetBytes32(in []byte) *Uint {
	_ = in[31] // bounds check hint to compiler; see golang.org/issue/14808
	z.arr[3] = binary.BigEndian.Uint64(in[0:8])
	z.arr[2] = binary.BigEndian.Uint64(in[8:16])
	z.arr[1] = binary.BigEndian.Uint64(in[16:24])
	z.arr[0] = binary.BigEndian.Uint64(in[24:32])
	return z
}

// Utility methods that are "missing" among the bigEndian.UintXX methods.

func bigEndianUint40(b []byte) uint64 {
	_ = b[4] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[4]) | uint64(b[3])<<8 | uint64(b[2])<<16 | uint64(b[1])<<24 |
		uint64(b[0])<<32
}

func bigEndianUint56(b []byte) uint64 {
	_ = b[6] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[6]) | uint64(b[5])<<8 | uint64(b[4])<<16 | uint64(b[3])<<24 |
		uint64(b[2])<<32 | uint64(b[1])<<40 | uint64(b[0])<<48
}

func bigEndianUint48(b []byte) uint64 {
	_ = b[5] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[5]) | uint64(b[4])<<8 | uint64(b[3])<<16 | uint64(b[2])<<24 |
		uint64(b[1])<<32 | uint64(b[0])<<40
}
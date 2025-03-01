package fuzzing

import (
	"encoding/binary"
)

const (
	prime64_1 = 11400714785074694791
	prime64_2 = 14029467366897019727
	prime64_3 = 1609587929392839161
	prime64_4 = 9650029242287828579
	prime64_5 = 2870177450012600261
)

func X_xxh64Sum(input []byte, seed uint64) uint64 {
	n := len(input)
	var h64 uint64

	// 1) set init accumulator

	var v1, v2, v3, v4 uint64
	idx := 0

	if n >= 32 {
		v1 = seed + prime64_1 + prime64_2
		v2 = seed + prime64_2
		v3 = seed
		v4 = seed - prime64_1

		limit := n - 32
		for idx <= limit {
			sub := input[idx : idx+32]
			v1 = rol31(v1+u64(sub[0:8])*prime64_2) * prime64_1
			v2 = rol31(v2+u64(sub[8:16])*prime64_2) * prime64_1
			v3 = rol31(v3+u64(sub[16:24])*prime64_2) * prime64_1
			v4 = rol31(v4+u64(sub[24:32])*prime64_2) * prime64_1
			idx += 32
		}

		h64 = rol1(v1) + rol7(v2) + rol12(v3) + rol18(v4)

		// merge
		v1 *= prime64_2
		v2 *= prime64_2
		v3 *= prime64_2
		v4 *= prime64_2

		h64 = (h64^(rol31(v1)*prime64_1))*prime64_1 + prime64_4
		h64 = (h64^(rol31(v2)*prime64_1))*prime64_1 + prime64_4
		h64 = (h64^(rol31(v3)*prime64_1))*prime64_1 + prime64_4
		h64 = (h64^(rol31(v4)*prime64_1))*prime64_1 + prime64_4

		h64 += uint64(n)
	} else {
		h64 = seed + prime64_5 + uint64(n)
	}

	// 2) remain byte(8 byte block)

	for n8 := n - 8; idx <= n8; idx += 8 {
		k := u64(input[idx : idx+8])
		h64 ^= rol31(k*prime64_2) * prime64_1
		h64 = rol27(h64)*prime64_1 + prime64_4
	}

	// 3) reamin byte (1~7)

	if (n - idx) >= 4 {
		k := binary.LittleEndian.Uint32(input[idx : idx+4])
		h64 ^= uint64(k) * prime64_1
		h64 = rol23(h64)*prime64_2 + prime64_3
		idx += 4
	}
	for ; idx < n; idx++ {
		h64 ^= uint64(input[idx]) * prime64_5
		h64 = rol11(h64) * prime64_1
	}

	// 4) last Avalanche

	h64 ^= h64 >> 33
	h64 *= prime64_2
	h64 ^= h64 >> 29
	h64 *= prime64_3
	h64 ^= h64 >> 32

	return h64
}

// rol1, rol7, rol11, rol12, rol18, rol23, rol27, rol31
func rol1(u uint64) uint64  { return (u << 1) | (u >> 63) }
func rol7(u uint64) uint64  { return (u << 7) | (u >> 57) }
func rol11(u uint64) uint64 { return (u << 11) | (u >> 53) }
func rol12(u uint64) uint64 { return (u << 12) | (u >> 52) }
func rol18(u uint64) uint64 { return (u << 18) | (u >> 46) }
func rol23(u uint64) uint64 { return (u << 23) | (u >> 41) }
func rol27(u uint64) uint64 { return (u << 27) | (u >> 37) }
func rol31(u uint64) uint64 { return (u << 31) | (u >> 33) }

// u64, u32
func u64(buf []byte) uint64 {
	// go compiler might optimize this pattern on little-endian
	return uint64(buf[0]) | uint64(buf[1])<<8 |
		uint64(buf[2])<<16 | uint64(buf[3])<<24 |
		uint64(buf[4])<<32 | uint64(buf[5])<<40 |
		uint64(buf[6])<<48 | uint64(buf[7])<<56
}

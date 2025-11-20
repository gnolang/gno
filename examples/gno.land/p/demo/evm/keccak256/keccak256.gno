package keccak256

import "encoding/binary"

const (
	rate      = 136 // 1088 bits for Keccak-256
	capacity  = 64  // 512 bits
	outputLen = 32  // 256 bits
)

var roundConstants = [24]uint64{
	0x0000000000000001, 0x0000000000008082, 0x800000000000808A,
	0x8000000080008000, 0x000000000000808B, 0x0000000080000001,
	0x8000000080008081, 0x8000000000008009, 0x000000000000008A,
	0x0000000000000088, 0x0000000080008009, 0x000000008000000A,
	0x000000008000808B, 0x800000000000008B, 0x8000000000008089,
	0x8000000000008003, 0x8000000000008002, 0x8000000000000080,
	0x000000000000800A, 0x800000008000000A, 0x8000000080008081,
	0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
}

var rotationOffsets = [5][5]int{
	{0, 36, 3, 41, 18},
	{1, 44, 10, 45, 2},
	{62, 6, 43, 15, 61},
	{28, 55, 25, 21, 56},
	{27, 20, 39, 8, 14},
}

func keccakF1600(state *[25]uint64) {
	var C [5]uint64
	var D [5]uint64
	var B [5][5]uint64

	for round := 0; round < 24; round++ {
		for x := 0; x < 5; x++ {
			C[x] = state[x] ^ state[x+5] ^ state[x+10] ^ state[x+15] ^ state[x+20]
		}
		for x := 0; x < 5; x++ {
			D[x] = C[(x+4)%5] ^ rotl(C[(x+1)%5], 1)
		}
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				state[x+5*y] ^= D[x]
			}
		}
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				B[y][(2*x+3*y)%5] = rotl(state[x+5*y], rotationOffsets[x][y])
			}
		}
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				state[x+5*y] = B[x][y] ^ (^B[(x+1)%5][y] & B[(x+2)%5][y])
			}
		}
		state[0] ^= roundConstants[round]
	}
}

func rotl(x uint64, n int) uint64 {
	return (x << n) | (x >> (64 - n))
}

// Perform the Keccak-256 hashing algorithm
func Keccak256(input []byte) []byte {
	var state [25]uint64
	var block [rate]byte

	padded := append([]byte{}, input...)
	padded = append(padded, 0x01)
	for len(padded)%rate != rate-1 {
		padded = append(padded, 0x00)
	}
	padded = append(padded, 0x80)
	for i := 0; i < len(padded); i += rate {
		copy(block[:], padded[i:i+rate])
		for j := 0; j < rate/8; j++ {
			state[j] ^= binary.LittleEndian.Uint64(block[j*8:])
		}
		keccakF1600(&state)
	}
	var output []byte
	for len(output) < outputLen {
		for i := 0; i < rate/8 && len(output) < outputLen; i++ {
			buf := make([]byte, 8)
			binary.LittleEndian.PutUint64(buf, state[i])
			output = append(output, buf...)
		}
		if len(output) < outputLen {
			keccakF1600(&state)
		}
	}

	return output[:outputLen]
}

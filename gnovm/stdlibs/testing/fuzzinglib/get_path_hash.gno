package fuzzinglib

import (
	"crypto/sha256"
	"strconv"
	"strings"
)

// Coverage 타입 정의
type Coverage []uint

// Set 구조체 정의 (uint64 전용)
type Set struct {
	data map[uint64]struct{}
}

type PathHash2Coverage_table struct {
	hash2Coverage map[uint64]Coverage
}

// func main() {
// 	case1 := Coverage{1, 11, 111}
// 	case2 := Coverage{11, 1, 111}
// 	case3 := Coverage{1}

// 	cp1 := Get_PathHash(case1)
// 	cp2 := Get_PathHash(case2)
// 	cp3 := Get_PathHash(case3)

// 	ec1 := Coverage{1, 11, 111}
// 	ec2 := Coverage{11, 1, 111}
// 	ec3 := Coverage{1}
// 	println(ec2.ToString())

// 	ep1 := Get_PathHash(ec1)
// 	ep2 := Get_PathHash(ec2)
// 	ep3 := Get_PathHash(ec3)

// 	println(cp1, ep1)
// 	println(cp1 == ep1)
// 	println(cp2, ep2)
// 	println(cp2 == ep2)
// 	println(cp3, ep3)
// 	println(cp3 == ep3)

// 	ph_table := New_PathHash2Coverage_table()
// 	ph_table.Add(cp1, case1)
// 	ph_table.Add(cp2, case2)
// 	ph_table.Add(cp3, case3)
// 	println(ph_table)

// 	ep_x := Get_PathHash(Coverage{111, 11, 1})
// 	println(ph_table.Contains(ep1))
// 	println(ph_table.Contains(ep_x))
// }

// operation function!!
func Get_PathHash(input Coverage) uint64 {
	valBytes := UintSliceToBytes(input)
	ValArray := sha256.Sum256(valBytes)
	return BytesToUint64(ValArray)
}

// ------------internal functions--------------
// Set 생성자
func New_PathHash2Coverage_table() *PathHash2Coverage_table {
	return &PathHash2Coverage_table{
		hash2Coverage: make(map[uint64]Coverage),
	}
}

// Set 메서드들
func (pct *PathHash2Coverage_table) Add(ph uint64, c Coverage) {
	pct.hash2Coverage[ph] = c
}

func (pct *PathHash2Coverage_table) Remove(ph uint64) {
	delete(pct.hash2Coverage, ph)
}

func (pct *PathHash2Coverage_table) Contains(ph uint64) bool {
	_, exists := pct.hash2Coverage[ph]
	return exists
}

func (pct *PathHash2Coverage_table) Size() int {
	return len(pct.hash2Coverage)
}

// uint 슬라이스를 문자열로 변환-deprecated
func (c Coverage) ToString() string {
	if len(c) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(strconv.FormatUint(uint64(c[0]), 10))

	for _, num := range c[1:] {
		builder.WriteByte(' ')
		builder.WriteString(strconv.FormatUint(uint64(num), 10))
	}

	return builder.String()
}

func UintSliceToBytes(uints []uint) []byte {
	result := make([]byte, len(uints)*8) // uint 하나당 8바이트

	for i, num := range uints {
		// uint 값을 8개의 byte로 변환
		for j := 0; j < 8; j++ {
			result[i*8+j] = byte(num >> uint(56-j*8))
		}
	}

	return result
}

// sha256이 uint8 32개 를 반환해서,,,
func BytesToUint64(b [32]uint8) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}

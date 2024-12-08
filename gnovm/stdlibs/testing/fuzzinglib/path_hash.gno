package fuzzinglib

import (
	"strconv"
	"strings"
)

// Coverage 타입 정의
type Coverage []uint

// Set 구조체 정의 (uint64 전용)
type Set struct {
	data map[uint64]struct{}
}

// Set 생성자
func NewSet() *Set {
	return &Set{
		data: make(map[uint64]struct{}),
	}
}

// Set 메서드들
func (s *Set) Add(element uint64) {
	s.data[element] = struct{}{}
}

func (s *Set) Remove(element uint64) {
	delete(s.data, element)
}

func (s *Set) Contains(element uint64) bool {
	_, exists := s.data[element]
	return exists
}

func (s *Set) Size() int {
	return len(s.data)
}

// uint 슬라이스를 문자열로 변환
func UintToString(nums []uint) string {
	if len(nums) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(strconv.FormatUint(uint64(nums[0]), 10))

	for _, num := range nums[1:] {
		builder.WriteByte(' ')
		builder.WriteString(strconv.FormatUint(uint64(num), 10))
	}

	return builder.String()
}

// Coverage 존재 여부 확인
func Is_exist_coverage(c Coverage, set *Set) bool {
	valToHash := UintToString(c)
	hash := UseSum64String(valToHash)
	return set.Contains(hash)
}

// UseSum64String 함수 정의
func UseSum64String(stVal string) uint64 {
	var result uint64
	for _, char := range stVal {
		result = result*31 + uint64(char)
	}
	return result
}

func Make_PathHash(c Coverage) uint64 {
	valToHash := UintToString(c)
	hash := UseSum64String(valToHash)
	return hash
}

// func main() {
// 	// 미리 정의된 set 만들어놓기
// 	coverageSet := NewSet()

// 	testVal1 := []uint{1, 11, 1337}
// 	testVal2 := []uint{11, 1, 1337}
// 	testVal3 := []uint{9832, 10, 1337}

// 	ph1 := Make_PathHash(testVal1)
// 	ph2 := Make_PathHash(testVal2)
// 	ph3 := Make_PathHash(testVal3)

// 	coverageSet.Add(ph1)
// 	coverageSet.Add(ph2)
// 	coverageSet.Add(ph3)

// 	// 테스트용 Val 생성
// 	testVal_1 := []uint{1, 11, 1337}
// 	t_ph1 := Make_PathHash(testVal_1)

// 	testVal_2 := []uint{11, 1, 1337}
// 	t_ph2 := Make_PathHash(testVal_2)

// 	testVal_3 := []uint{9832, 10, 1337}
// 	t_ph1 := Make_PathHash(1testVal_3)

// 	println("equal_inp->", ph1, "|", t_ph1)
// 	println("equal_inp->", ph2, "|",  t_ph2)
// 	println("equal_inp->", ph3, "|",  t_ph3)
// }

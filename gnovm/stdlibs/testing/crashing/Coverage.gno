package crashing

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

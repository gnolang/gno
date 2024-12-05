package main

import (
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
)

// Coverage 타입 정의
type Coverage []uint

// 제네릭 Set 구조체 정의
type Set[T comparable] struct {
	data map[T]struct{}
}

// Set 생성자
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		data: make(map[T]struct{}),
	}
}

// Set 메서드들
func (s *Set[T]) Add(element T) {
	s.data[element] = struct{}{}
}

func (s *Set[T]) Remove(element T) {
	delete(s.data, element)
}

func (s *Set[T]) Contains(element T) bool {
	_, exists := s.data[element]
	return exists
}

func (s *Set[T]) Size() int {
	return len(s.data)
}

// uint 슬라이스를 문자열로 변환
func uintToString(nums []uint) string {
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
func Is_exist_coverage(c Coverage, set *Set[uint64]) bool {
	valToHash := uintToString(c)
	hash := xxhash.Sum64String(valToHash)
	return set.Contains(hash)
}

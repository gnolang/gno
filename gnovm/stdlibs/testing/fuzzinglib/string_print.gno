package fuzzinglib

import (
	"strconv"
	"strings"
)

// uint 값을 문자열로 변환
func uintToString(v uint) string {
	return strconv.Itoa(int(v))
}

// bool 값을 문자열로 변환
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// []uint 슬라이스를 문자열로 변환
func sliceToString(slice []uint) string {
	if len(slice) == 0 {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i, val := range slice {
		sb.WriteString(uintToString(val))
		if i < len(slice)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// []Seed 슬라이스를 문자열로 변환
func seedsToString(seeds []Seed) string {
	if len(seeds) == 0 {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteString("[\n")
	for _, seed := range seeds {
		sb.WriteString("    {Gen: ")
		sb.WriteString(uintToString(seed.Gen))
		sb.WriteString(", IsUsed: ")
		sb.WriteString(boolToString(seed.IsUsed))
		sb.WriteString(", Content: ")
		sb.WriteString(seed.Content)
		sb.WriteString(", Result: ")
		sb.WriteString(interfaceToString(seed.Result))
		sb.WriteString("},\n")
	}
	sb.WriteString("]")
	return sb.String()
}

// interface{} 값을 문자열로 변환
func interfaceToString(i interface{}) string {
	switch v := i.(type) {
	case nil:
		return "nil"
	case string:
		return strconv.Quote(v) // 문자열을 따옴표로 감쌉니다.
	case int, int32, int64:
		return strconv.FormatInt(int64(v.(int)), 10)
	case uint, uint32, uint64:
		return strconv.FormatUint(uint64(v.(uint)), 10)
	case float32, float64:
		return strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return "unknown"
	}
}

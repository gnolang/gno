package testing

import (
	"bytes"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Token은 “정상 유니코드”인지, “깨진 바이트”인지 구분.
type Token struct {
	Data  []byte // 토큰의 실제 바이트
	Valid bool   // true이면 UTF-8 디코딩에 성공한 문자
}

// tokenizeString: 문자열 s를 가능한 한 UTF-8로 디코딩하고, 깨진 바이트는 따로 보관
func TokenizeString(s string) []Token {
	var tokens []Token
	b := []byte(s)
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			// 디코딩 실패 → 깨진 바이트 하나를 토큰으로
			tokens = append(tokens, Token{
				Data:  []byte{b[i]},
				Valid: false,
			})
			i++
		default:
			// 디코딩 성공 (유니코드 문자 하나)
			tokens = append(tokens, Token{
				Data:  b[i : i+size],
				Valid: true,
			})
			i += size
		}
	}
	return tokens
}

func RebuildString(tokens []Token) string {
	var buf bytes.Buffer
	for _, t := range tokens {
		buf.Write(t.Data) // 토큰의 바이트 시퀀스를 그대로 연결
	}
	return buf.String()
}

func RebuildEscaped(tokens []Token) string {
	var result []byte // 최종 바이트를 담을 슬라이스
	for _, t := range tokens {
		if t.Valid {
			// 유효한 UTF-8 문자이므로 그대로 추가
			result = append(result, t.Data...)
		} else {
			// 깨진 바이트 -> \xNN 형태로 변환
			// (질문 내용대로, 보통 invalid token은 1바이트씩 들어있음)
			for _, b := range t.Data {
				result = append(result, []byte("\\x")...) // "\x"
				hex := byteToHex(b)                       // 예: 0xEB -> "eb"
				result = append(result, hex...)
			}
		}
	}
	return string(result)
}

// byteToHex: 한 바이트를 2자리 16진수 문자열로 반환 (예: 0xAB -> "ab")
func byteToHex(b byte) []byte {
	const hexdigits = "0123456789abcdef"
	hi := hexdigits[b>>4]
	lo := hexdigits[b&0x0F]
	return []byte{hi, lo} // 예: [ 'a', 'b' ]
}

func TransForHuman(s string) string {
	toks := TokenizeString(s)
	escaped := RebuildEscaped(toks)
	return escaped
}

// ///////////////////
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

func interfacesliceToString(slice []interface{}) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, elem := range slice {
		sb.WriteString(interfaceToString(elem)) // 각 요소를 문자열로 변환
		if i < len(slice)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// coverageToString: Coverage 타입을 문자열로 변환
func CoverageToString(coverage Coverage) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, line := range coverage {
		sb.WriteString("{co_name: ")
		sb.WriteString(line.co_name)
		sb.WriteString(", co_line: ")
		sb.WriteString(strconv.Itoa(line.co_line)) // fmt.Sprintf 대신 strconv 사용
		sb.WriteString("}")
		if i < len(coverage)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func ContentToString(content []interface{}) string {
	var result strings.Builder

	result.WriteString("[")
	for i, elem := range content {
		switch v := elem.(type) {
		case string:
			// 문자열은 따옴표로 감쌈
			result.WriteString("\"" + v + "\"")
		case int:
			// int를 문자열로 변환
			result.WriteString(strconv.Itoa(v))
		case int8:
			// int8을 int64로 변환 후 문자열로 변환
			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int16:
			// int16을 int64로 변환 후 문자열로 변환
			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int32:
			// int32를 int64로 변환 후 문자열로 변환
			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int64:
			// int64를 문자열로 변환
			result.WriteString(strconv.FormatInt(v, 10))
		case uint:
			// uint를 uint64로 변환 후 문자열로 변환
			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint8:
			// uint8을 uint64로 변환 후 문자열로 변환
			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint16:
			// uint16을 uint64로 변환 후 문자열로 변환
			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint32:
			// uint32를 uint64로 변환 후 문자열로 변환
			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint64:
			// uint64를 문자열로 변환
			result.WriteString(strconv.FormatUint(v, 10))
		case float32:
			// float32를 문자열로 변환
			result.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
		case float64:
			// float64를 문자열로 변환
			result.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		case []byte:
			// []byte를 문자열로 변환
			result.WriteString("\"" + string(v) + "\"")
		case bool:
			// bool 값을 true/false 문자열로 변환
			if v {
				result.WriteString("true")
			} else {
				result.WriteString("false")
			}
		default:
			// 지원되지 않는 타입은 "unknown"으로 처리
			result.WriteString("unknown")
		}

		// 마지막 요소가 아니면 콤마 추가
		if i < len(content)-1 {
			result.WriteString(", ")
		}
	}
	result.WriteString("]")

	return result.String()
}

// func SeedToString(seed Seed) string {

// }

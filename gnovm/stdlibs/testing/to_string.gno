package testing

import (
	"bytes"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Token distinguishes whether it is "normal Unicode" or "broken bytes".
type Token struct {
	Data  []byte // Actual bytes of the token
	Valid bool   // If true, characters that have successfully decoded UTF-8
}

// tokenizeString: Decode string s as UTF-8 as much as possible, keep broken bytes separate
func TokenizeString(s string) []Token {
	var tokens []Token
	b := []byte(s)
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			tokens = append(tokens, Token{
				Data:  []byte{b[i]},
				Valid: false,
			})
			i++
		default:
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
		buf.Write(t.Data)
	}
	return buf.String()
}

func RebuildEscaped(tokens []Token) string {
	var result []byte
	for _, t := range tokens {
		if t.Valid {
			result = append(result, t.Data...)
		} else {
			for _, b := range t.Data {
				result = append(result, []byte("\\x")...)
				hex := byteToHex(b)
				result = append(result, hex...)
			}
		}
	}
	return string(result)
}

func byteToHex(b byte) []byte {
	const hexdigits = "0123456789abcdef"
	hi := hexdigits[b>>4]
	lo := hexdigits[b&0x0F]
	return []byte{hi, lo}
}

func TransForHuman(s string) string {
	toks := TokenizeString(s)
	escaped := RebuildEscaped(toks)
	return escaped
}

func uintToString(v uint) string {
	return strconv.Itoa(int(v))
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

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

func interfaceToString(i interface{}) string {
	switch v := i.(type) {
	case nil:
		return "nil"
	case string:
		return strconv.Quote(v)
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
		sb.WriteString(interfaceToString(elem))
		if i < len(slice)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func CoverageToString(coverage Coverage) string {
	var sb strings.Builder
	sb.WriteString("[")
	count := 0
	for i, line := range coverage {
		sb.WriteString("{co_name: ")
		sb.WriteString(line.co_name)
		sb.WriteString(", co_line: ")
		sb.WriteString(strconv.Itoa(line.co_line))
		sb.WriteString("}")
		count++
		if i < len(coverage)-1 {
			if count%3 == 0 {
				sb.WriteString(", \n")
			} else {
				sb.WriteString(", ")
			}
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
			result.WriteString("\"" + v + "\"")
		case int:

			result.WriteString(strconv.Itoa(v))
		case int8:

			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int16:

			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int32:

			result.WriteString(strconv.FormatInt(int64(v), 10))
		case int64:

			result.WriteString(strconv.FormatInt(v, 10))
		case uint:

			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint8:

			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint16:

			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint32:

			result.WriteString(strconv.FormatUint(uint64(v), 10))
		case uint64:

			result.WriteString(strconv.FormatUint(v, 10))
		case float32:

			result.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
		case float64:

			result.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		case []byte:

			result.WriteString("\"" + string(v) + "\"")
		case bool:

			if v {
				result.WriteString("true")
			} else {
				result.WriteString("false")
			}
		default:

			result.WriteString("unknown")
		}

		if i < len(content)-1 {
			result.WriteString(", ")
		}
	}
	result.WriteString("]")

	return result.String()
}

func CrashCaseToString(c Crash_Case) string {
	var sb strings.Builder

	sb.WriteString("{ \n")

	sb.WriteString("Input: ")
	sb.WriteString(TransForHuman(c.Input))
	if c.IsPanic {
		sb.WriteString("\nPanic/Error: Panic ")
		sb.WriteString("\nPanicMessage: \"")
		sb.WriteString(c.PanicMessage)
	} else {
		sb.WriteString("\nPanic/Error: Error ")
		sb.WriteString("\nErrorMessage: ")
		sb.WriteString("\"" + c.ErrorMsg + "\"")
	}

	sb.WriteString("\n}")

	return sb.String()
}

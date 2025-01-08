package testing

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

type (
	fuzzLogger struct {
		data []crashCase
	}
	crashCase struct {
		hashNumber   uint
		input        string
		rawContent   []interface{}
		inputCount   uint
		isPanic      bool
		panicMessage string
		errorMsg     string
	}
)

type testResult struct {
	panicOccurred bool
	panicMessage  string
	prror         error
}

func (fl *fuzzLogger) addCase(seed seed, r testResult) {
	var crashcase crashCase
	if r.prror == nil {
		crashcase = crashCase{
			hashNumber:   uint(seed.hashNumber),
			input:        contentToString(seed.content),
			rawContent:   seed.content,
			inputCount:   seed.id,
			isPanic:      r.panicOccurred,
			panicMessage: r.panicMessage,
			errorMsg:     "",
		}
	} else {
		crashcase = crashCase{
			hashNumber:   uint(seed.hashNumber),
			input:        contentToString(seed.content),
			rawContent:   seed.content,
			inputCount:   seed.id,
			isPanic:      r.panicOccurred,
			panicMessage: r.panicMessage,
			errorMsg:     strings.TrimSpace(r.prror.Error()),
		}
	}
	fl.data = append(fl.data, crashcase)
}

func (fl *fuzzLogger) reportLastFailingInput() {
	log := fl.data
	lastCrashCase := log[len(log)-1]
	fmt.Fprintf(os.Stderr, "Found failing input %s at %d trials\n", escapedToString(lastCrashCase.input), lastCrashCase.inputCount)
	return
}

// Recover what the escape characters break when they are printed. (e.g."�" -> "\xeb")
func escapedToString(s string) string {
	toks := tokenizeString(s)
	escaped := rebuildEscaped(toks)
	return escaped
}

// token distinguishes whether it is "normal Unicode" or "broken bytes".
type token struct {
	Data  []byte // Actual bytes of the token
	Valid bool   // If true, characters that have successfully decoded UTF-8
}

// tokenizeString: Decode string s as UTF-8 as much as possible, keep broken bytes separate
func tokenizeString(s string) []token {
	var tokens []token
	b := []byte(s)
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			tokens = append(tokens, token{
				Data:  []byte{b[i]},
				Valid: false,
			})
			i++
		default:
			tokens = append(tokens, token{
				Data:  b[i : i+size],
				Valid: true,
			})
			i += size
		}
	}
	return tokens
}

func rebuildString(tokens []token) string {
	var buf bytes.Buffer
	for _, t := range tokens {
		buf.Write(t.Data)
	}
	return buf.String()
}

func rebuildEscaped(tokens []token) string {
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

func (fl *fuzzLogger) reportLastCrashCase() {
	log := fl.data
	lastCrashCase := log[len(log)-1]
	fmt.Fprintf(os.Stderr, "%s\n", crashCaseToString(lastCrashCase))
	return
}

func crashCaseToString(c crashCase) string {
	var sb strings.Builder

	sb.WriteString("Input: ")
	sb.WriteString(escapedToString(c.input))
	if c.isPanic {
		sb.WriteString("\nPanic/Error: Panic ")
		sb.WriteString("\nPanicMessage: \"")
		sb.WriteString(c.panicMessage)
	} else {
		sb.WriteString("\nPanic/Error: Error ")
		sb.WriteString("\nErrorMessage: ")
		sb.WriteString("\"" + c.errorMsg + "\"")
	}

	return sb.String()
}

func contentToString(content []interface{}) string {
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

func (fl *fuzzLogger) getLastCrashCase() crashCase {
	log := fl.data
	lastCrashCase := log[len(log)-1]
	return lastCrashCase
}

func newFuzzLogger() *fuzzLogger {
	return &fuzzLogger{
		data: make([]crashCase, 0),
	}
}

package testing

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// fuzzLogger records and manages failure cases during fuzzing.
type fuzzLogger struct {
	data []crashCase
}

// crashCase represents a single failure case encountered during fuzzing.
type crashCase struct {
	hashNumber   uint          // The unique hash number associated with the coverage.
	input        string        // The input string that caused the failure.
	rawContent   []interface{} // The raw content of the input.
	inputCount   uint          // The count of inputs processed before this failure.
	isPanic      bool          // Indicates whether a panic occurred.
	panicMessage string        // The panic message, if a panic occurred.
	errorMsg     string        // The error message, if an error occurred.
}

// testResult holds the result of a fuzzing test, indicating if a panic occurred and any associated messages.
type testResult struct {
	panicOccurred bool   // Indicates whether a panic occurred during the test.
	panicMessage  string // The message associated with the panic.
	prror         error  // The error encountered, if any.
}

// addCase adds a new crash case to the fuzzLogger based on the provided seed and testResult.
// It assigns a unique hash number to the coverage and records the details of the failure.
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

// reportLastFailingInput reports the most recent failing input to standard error.
func (fl *fuzzLogger) reportLastFailingInput() {
	log := fl.data
	lastCrashCase := log[len(log)-1]
	fmt.Fprintf(os.Stderr, "Found failing input %s at %d trials\n", escapedToString(lastCrashCase.input), lastCrashCase.inputCount)
	return
}

// escapedToString converts escaped characters back to their original form when printed.
// For example, it converts "�" to "\xeb".
func escapedToString(s string) string {
	toks := tokenizeString(s)
	escaped := rebuildEscaped(toks)
	return escaped
}

// token represents a segment of a string, distinguishing between valid Unicode and broken bytes.
type token struct {
	Data  []byte // The actual bytes of the token.
	Valid bool   // Indicates if the bytes successfully decoded as UTF-8.
}

// tokenizeString decodes the input string as UTF-8 as much as possible, separating broken bytes.
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

// rebuildString reconstructs a string from a slice of tokens.
func rebuildString(tokens []token) string {
	var buf bytes.Buffer
	for _, t := range tokens {
		buf.Write(t.Data)
	}
	return buf.String()
}

// rebuildEscaped reconstructs a string with escaped bytes from a slice of tokens.
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

// byteToHex converts a byte to its hexadecimal representation.
func byteToHex(b byte) []byte {
	const hexdigits = "0123456789abcdef"
	hi := hexdigits[b>>4]
	lo := hexdigits[b&0x0F]
	return []byte{hi, lo}
}

// reportLastCrashCase reports the details of the most recent crash case to standard error.
func (fl *fuzzLogger) reportLastCrashCase() {
	log := fl.data
	lastCrashCase := log[len(log)-1]
	fmt.Fprintf(os.Stderr, "%s\n", crashCaseToString(lastCrashCase))
	return
}

// crashCaseToString formats a crashCase into a readable string representation.
func crashCaseToString(c crashCase) string {
	var sb strings.Builder

	sb.WriteString("Input: ")
	sb.WriteString(escapedToString(c.input))
	if c.isPanic {
		sb.WriteString("\nPanic/Error: Panic")
		sb.WriteString("\nPanicMessage: \"")
		sb.WriteString(c.panicMessage)
		sb.WriteString("\"")
	} else {
		sb.WriteString("\nPanic/Error: Error")
		sb.WriteString("\nErrorMessage: ")
		sb.WriteString("\"" + c.errorMsg + "\"")
	}

	return sb.String()
}

// contentToString converts a slice of interface{} content to its string representation.
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

// getLastCrashCase retrieves the most recent crash case from the fuzzLogger.
func (fl *fuzzLogger) getLastCrashCase() crashCase {
	log := fl.data
	lastCrashCase := log[len(log)-1]
	return lastCrashCase
}

// newFuzzLogger creates and initializes a new instance of fuzzLogger.
func newFuzzLogger() *fuzzLogger {
	return &fuzzLogger{
		data: make([]crashCase, 0),
	}
}

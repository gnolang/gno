package fuzzing

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// FuzzLogger records and manages failure cases during fuzzing.
type FuzzLogger struct {
	Data []CrashCase
}

// CrashCase represents a single failure case encountered during fuzzing.
type CrashCase struct {
	HashNumber   uint          // The unique hash number associated with the coverage.
	Input        string        // The input string that caused the failure.
	RawContent   []interface{} // The raw content of the input.
	InputCount   uint          // The count of inputs processed before this failure.
	IsPanic      bool          // Indicates whether a panic occurred.
	PanicMessage string        // The panic message, if a panic occurred.
	ErrorMsg     string        // The error message, if an error occurred.
}

// TestResult holds the result of a fuzzing test, indicating if a panic occurred and any associated messages.
type TestResult struct {
	PanicOccurred bool   // Indicates whether a panic occurred during the test.
	PanicMessage  string // The message associated with the panic.
	Error         error  // The error encountered, if any.
}

// addCase adds a new crash case to the fuzzLogger based on the provided seed and testResult.
// It assigns a unique hash number to the coverage and records the details of the failure.
func (fl *FuzzLogger) AddCase(seed Seed, r TestResult) {
	var crashcase CrashCase
	if r.Error == nil {
		crashcase = CrashCase{
			HashNumber:   uint(seed.HashNumber),
			Input:        ContentToString(seed.Content),
			RawContent:   seed.Content,
			InputCount:   seed.Id,
			IsPanic:      r.PanicOccurred,
			PanicMessage: r.PanicMessage,
			ErrorMsg:     "",
		}
	} else {
		crashcase = CrashCase{
			HashNumber:   uint(seed.HashNumber),
			Input:        ContentToString(seed.Content),
			RawContent:   seed.Content,
			InputCount:   seed.Id,
			IsPanic:      r.PanicOccurred,
			PanicMessage: r.PanicMessage,
			ErrorMsg:     strings.TrimSpace(r.Error.Error()),
		}
	}
	fl.Data = append(fl.Data, crashcase)
}

// reportLastFailingInput reports the most recent failing input to standard error.
func (fl *FuzzLogger) ReportLastFailingInput() {
	log := fl.Data
	lastCrashCase := log[len(log)-1]
	fmt.Fprintf(os.Stderr, "Found failing input %s at %d trials, coverage %d \n", escapedToString(lastCrashCase.Input), lastCrashCase.InputCount, int(lastCrashCase.HashNumber))
	return
}

// escapedToString converts escaped characters back to their original form when printed.
// For example, it converts "�" to "\xeb".
func escapedToString(s string) string {
	toks := TokenizeXXX(s)
	escaped := RebuildEscaped(toks)
	return escaped
}

func EscapedToString(s string) string {
	return escapedToString(s)
}

// Token represents a segment of a string, distinguishing between valid Unicode and broken bytes.
type Token struct {
	Data  []byte // The actual bytes of the token.
	Valid bool   // Indicates if the bytes successfully decoded as UTF-8.
}

func TokenizeXXX(s interface{}) []Token {
	var tokens []Token
	var b []byte

	switch v := s.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		panic("unsupported type: must be string or []byte")
	}
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

// RebuildString reconstructs a string from a slice of tokens.
func RebuildString(tokens []Token) string {
	var buf bytes.Buffer
	for _, t := range tokens {
		buf.Write(t.Data)
	}
	return buf.String()
}

// RebuildBytes: Converts the `tokens` slice back to `[]byte`
func RebuildBytes(tokens []Token) []byte {
	var buf bytes.Buffer
	for _, t := range tokens {
		buf.Write(t.Data) // Adds `token.Data` ([]byte) as it is
	}
	return buf.Bytes()
}

// RebuildEscaped reconstructs a string with escaped bytes from a slice of tokens.
func RebuildEscaped(tokens []Token) string {
	var result []byte
	for _, t := range tokens {
		if t.Valid {
			result = append(result, t.Data...)
		} else {
			for _, b := range t.Data {
				result = append(result, []byte("\\x")...)
				hex := ByteToHex(b)
				result = append(result, hex...)
			}
		}
	}
	return string(result)
}

// ByteToHex converts a byte to its hexadecimal representation.
func ByteToHex(b byte) []byte {
	const hexdigits = "0123456789abcdef"
	hi := hexdigits[b>>4]
	lo := hexdigits[b&0x0F]
	return []byte{hi, lo}
}

// reportLastCrashCase reports the details of the most recent crash case to standard error.
func (fl *FuzzLogger) ReportLastCrashCase() {
	log := fl.Data
	lastCrashCase := log[len(log)-1]
	fmt.Fprintf(os.Stderr, "%s\n", CrashCaseToString(lastCrashCase))
	return
}

// CrashCaseToString formats a crashCase into a readable string representation.
func CrashCaseToString(c CrashCase) string {
	var sb strings.Builder

	sb.WriteString("Input: ")
	sb.WriteString(escapedToString(c.Input))
	if c.IsPanic {
		sb.WriteString("\nPanic/Error: Panic")
		sb.WriteString("\nPanicMessage: \"")
		sb.WriteString(c.PanicMessage)
		sb.WriteString("\"")
	} else {
		sb.WriteString("\nPanic/Error: Error")
		sb.WriteString("\nErrorMessage: ")
		sb.WriteString("\"" + c.ErrorMsg + "\"")
	}

	return sb.String()
}

// ContentToString converts a slice of interface{} content to its string representation.
func ContentToString(content []interface{}) string {
	var result strings.Builder

	result.WriteString("[")
	for i, elem := range content {
		switch v := elem.(type) {
		case string:
			result.WriteString("\"" + escapedToString(v) + "\"")
		case []byte:
			result.WriteString("\"" + escapedToString(string(v)) + "\"")
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

// LastCrashCase retrieves the most recent crash case from the fuzzLogger.
func (fl *FuzzLogger) LastCrashCase() CrashCase {
	log := fl.Data
	lastCrashCase := log[len(log)-1]
	return lastCrashCase
}

// NewFuzzLogger creates and initializes a new instance of fuzzLogger.
func NewFuzzLogger() *FuzzLogger {
	return &FuzzLogger{
		Data: make([]CrashCase, 0),
	}
}

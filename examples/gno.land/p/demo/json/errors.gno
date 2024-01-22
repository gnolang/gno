package json

import "errors"

const (
	// KeyPathNotFoundError occurs when a specified key path does not exist in the JSON structure.
	KeyPathNotFoundError error = errors.New("JSON Error: key path not found in the JSON structure")

	// ArrayIndexNotFound occurs when the specified index is beyond the range of the array.
	ArrayIndexNotFound = errors.New("JSON Error: array index not found or out of range")

	// TokenNotFound occurs when a particular token (expected as part of the structure) is not found.
	TokenNotFound = errors.New("JSON Error: expected token not found in the JSON input")

	// KeyLevelNotMatched occurs when the key levels do not match the expected structure or depth.
	KeyLevelNotMatched = errors.New("JSON Error: key level does not match the expected structure or depth")

	// Overflow occurs when a number in the JSON exceeds the range that can be handled.
	Overflow = errors.New("JSON Error: numeric value exceeds the range limit")

	// EmptyBytes occurs when the JSON input is empty or has no content.
	EmptyBytes = errors.New("JSON Error: empty bytes: the JSON input is empty or has no content")

	// InvalidArrayIndex occurs when the index used for an array is not an integer or out of valid range.
	InvalidArrayIndex = errors.New("JSON Error: invalid array index: index should be an integer and within the valid range")

	// InvalidExponents occurs when there's an error related to the format or range of exponents in numbers.
	InvalidExponents = errors.New("JSON Error: invalid format or range of exponents in a numeric value")

	// NonDigitCharacters occurs when there are non-digit characters where a number is expected.
	NonDigitCharacters = errors.New("JSON Error: non-digit characters found where a number is expected")

	// MultipleDecimalPoints occurs when a number has more than one decimal point.
	MultipleDecimalPoints = errors.New("JSON Error: multiple decimal points found in a number")

	// MalformedType occurs when the type of a value does not match the expected type.
	MalformedType = errors.New("JSON Error: malformed type: the type of the value does not match the expected type")

	// MalformedString occurs when a string is improperly formatted, like unescaped characters or incorrect quotes.
	MalformedString = errors.New("JSON Error: malformed string: improperly formatted string, check for unescaped characters or incorrect quotes")

	// MalformedValue occurs when a value does not conform to the expected format or structure.
	MalformedValue = errors.New("JSON Error: malformed value: the value does not conform to the expected format or structure")

	// MalformedObject occurs when a JSON object is improperly formatted.
	MalformedObject = errors.New("JSON Error: malformed object: the JSON object is improperly formatted or structured")

	// MalformedArray occurs when a JSON array is improperly formatted.
	MalformedArray = errors.New("JSON Error: malformed array: the JSON array is improperly formatted or structured")

	// MalformedJson occurs when the entire JSON structure is improperly formatted or structured.
	MalformedJson = errors.New("JSON Error: malformed JSON: the entire JSON structure is improperly formatted or structured")

	// UnknownValueType occurs when the JSON contains a value of an unrecognized or unsupported type.
	UnknownValueType = errors.New("JSON Error: unknown value type: the value type is unrecognized or unsupported")
)

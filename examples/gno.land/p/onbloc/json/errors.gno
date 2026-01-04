package json

import "errors"

var (
	errNilNode               = errors.New("node is nil")
	errNotArrayNode          = errors.New("node is not array")
	errNotBoolNode           = errors.New("node is not boolean")
	errNotNullNode           = errors.New("node is not null")
	errNotNumberNode         = errors.New("node is not number")
	errNotObjectNode         = errors.New("node is not object")
	errNotStringNode         = errors.New("node is not string")
	errInvalidToken          = errors.New("invalid token")
	errIndexNotFound         = errors.New("index not found")
	errInvalidAppend         = errors.New("can't append value to non-appendable node")
	errInvalidAppendCycle    = errors.New("appending value to itself or its children or parents will cause a cycle")
	errInvalidEscapeSequence = errors.New("invalid escape sequence")
	errInvalidStringValue    = errors.New("invalid string value")
	errEmptyBooleanNode      = errors.New("boolean node is empty")
	errEmptyStringNode       = errors.New("string node is empty")
	errKeyRequired           = errors.New("key is required for object")
	errUnmatchedParenthesis  = errors.New("mismatched bracket or parenthesis")
	errUnmatchedQuotePath    = errors.New("unmatched quote in path")
)

var (
	errInvalidStringInput    = errors.New("invalid string input")
	errMalformedBooleanValue = errors.New("malformed boolean value")
	errEmptyByteSlice        = errors.New("empty byte slice")
	errInvalidExponentValue  = errors.New("invalid exponent value")
	errNonDigitCharacters    = errors.New("non-digit characters found")
	errNumericRangeExceeded  = errors.New("numeric value exceeds the range limit")
	errMultipleDecimalPoints = errors.New("multiple decimal points found")
)

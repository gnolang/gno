package json

import "errors"

const (
	KeyPathNotFoundError  error = errors.New("key path not found")
	ArrayIndexNotFound  	    = errors.New("array index not found")
	NotSpecifiedKeyError        = errors.New("key not specified")
	ClosingBracketError         = errors.New("closing bracket not found")
	ClosingQuoteError           = errors.New("closing quote not found")
	StringNotClosed             = errors.New("string not closed")
	BlockIsNotClosed            = errors.New("block is not closed")
	TokenNotFound               = errors.New("token not found")
	OutOfLevel                  = errors.New("out of level")
	KeyLevelNotMatched          = errors.New("key level not matched")
	Overflow                    = errors.New("overflow")
	EmptyBytes                  = errors.New("empty bytes")
	InvalidArrayIndex		    = errors.New("invalid array index")
	InvalidExponents            = errors.New("invalid exponents")
	NonDigitCharacters          = errors.New("non-digit characters")
	MultipleDecimalPoints       = errors.New("multiple decimal points")
	MalformedStringError        = errors.New("malformed string")
	MalformedValueError         = errors.New("malformed value")
	MalformedObjectError        = errors.New("malformed object")
	MalformedArrayError         = errors.New("malformed array")
	MalformedJsonError          = errors.New("malformed json array")
	UnknownValueTypeError       = errors.New("unknown value type")
)

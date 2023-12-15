package json

import "errors"

const (
	KeyPathNotFoundError  error = errors.New("key path not found")
	ArrayIndexNotFound  	    = errors.New("array index not found")
	TokenNotFound               = errors.New("token not found")
	KeyLevelNotMatched          = errors.New("key level not matched")
	Overflow                    = errors.New("overflow")
	EmptyBytes                  = errors.New("empty bytes")
	InvalidArrayIndex		    = errors.New("invalid array index")
	InvalidExponents            = errors.New("invalid exponents")
	NonDigitCharacters          = errors.New("non-digit characters")
	MultipleDecimalPoints       = errors.New("multiple decimal points")
	MalformedString       		= errors.New("malformed string")
	MalformedValue     		    = errors.New("malformed value")
	MalformedObject    		    = errors.New("malformed object")
	MalformedArray      		= errors.New("malformed array")
	MalformedJson      		    = errors.New("malformed json array")
	UnknownValueType  		    = errors.New("unknown value type")
)

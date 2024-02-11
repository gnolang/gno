package json

import "regexp"

const (
	supplementalPlanesOffset = 0x10000
	highSurrogateOffset      = 0xD800
	lowSurrogateOffset       = 0xDC00

	surrogateEnd                 = 0xDFFF
	basicMultilingualPlaneOffset = 0xFFFF

	badHex = -1

	unescapeStackBufSize = 64
	float32ExponentBias = 127
	float64ExponentBias = 1023
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = 1<<63 - 1
	maxUint64   = 1<<64 - 1
	intSize     = 32 << (^uint(0) >> 63)
	IntSize     = intSize
)

var (
	tabRegex           = regexp.MustCompile(`(?m)^\t+`)
	lastLineSpaceRegex = regexp.MustCompile(`(?m)^\s+\}`)
	jsonKeyRegex       = regexp.MustCompile(`"([^"\\]|\\.)*"\s*:`)
)

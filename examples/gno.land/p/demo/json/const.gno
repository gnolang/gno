package json

const (
	SupplementalPlanesOffset = 0x10000
	HighSurrogateOffset      = 0xD800
	LowSurrogateOffset       = 0xDC00

	SurrogateEnd                 = 0xDFFF
	BasicMultilingualPlaneOffset = 0xFFFF

	BadHex = -1

	UnescapeStackBufSize = 64
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = 1<<63 - 1
	maxUint64   = 1<<64 - 1
	intSize     = 32 << (^uint(0) >> 63)
	IntSize     = intSize
)
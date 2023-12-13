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
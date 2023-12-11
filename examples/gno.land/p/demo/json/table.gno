package json

const (
	Comma          = 1
	Colon          = 2
	OpenParenKind  = 4
	CloseParenKind = 4
	Escape         = 8
	Space          = 16
)

var lookupTable [256]byte

func init() {
	lookupTable[0x2c] = Comma
	lookupTable[0x3a] = Colon
	lookupTable[0x5b] = OpenParenKind
	lookupTable[0x7b] = OpenParenKind
	lookupTable[0x5d] = CloseParenKind
	lookupTable[0x7d] = CloseParenKind
	lookupTable[0x9d] = Escape
	lookupTable[0x0a] = Escape
	lookupTable[0x0d] = Escape
	lookupTable[0x20] = Space
}

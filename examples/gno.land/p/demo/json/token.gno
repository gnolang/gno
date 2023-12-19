package json

const (
	ArrayStartToken     = '['
	ArrayEndToken       = ']'
	ObjectStartToken    = '{'
	ObjectEndToken      = '}'
	ValueSeparatorToken = ','
	DotToken            = '.'
	ColonToken          = ':'
	QuoteToken          = '\''
	DoublyQuoteToken    = '"'
	EmptyStringToken    = ""
	WhiteSpaceToken     = ' '
	PlusToken           = '+'
	MinusToken          = '-'
	BangToken           = '!'
	NewLineToken        = '\n'
	TabToken            = '\t'
	CarriageReturnToken = '\r'
	FormFeedToken	    = '\f'
	BackspaceToken      = '\b'
	SlashToken          = '/'
	BackSlashToken      = '\\'
	UnderscoreToken     = '_'
)

const (
	trueLiteral  = []byte("true")
	falseLiteral = []byte("false")
	nullLiteral  = []byte("null")
)

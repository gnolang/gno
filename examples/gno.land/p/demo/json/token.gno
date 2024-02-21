package json

const (
	SquareOpenToken     = '['
	SquareCloseToken    = ']'
	CurlyOpenToken      = '{'
	CurlyCloseToken     = '}'
	RoundOpenToken      = '('
	RoundCloseToken     = ')'
	CommaToken          = ','
	DotToken            = '.'
	ColonToken          = ':'
	BackTickToken       = '`'
	QuoteToken          = '\''
	DoublyQuoteToken    = '"'
	EmptyStringToken    = ""
	WhiteSpaceToken     = ' '
	PlusToken           = '+'
	MinusToken          = '-'
	AesteriskToken      = '*'
	BangToken           = '!'
	QuestionToken       = '?'
	NewLineToken        = '\n'
	TabToken            = '\t'
	CarriageReturnToken = '\r'
	FormFeedToken       = '\f'
	BackSpaceToken      = '\b'
	SlashToken          = '/'
	BackSlashToken      = '\\'
	UnderScoreToken     = '_'
	DollarToken         = '$'
	AtToken             = '@'
	AndToken			= '&'
	OrToken				= '|'
)

const (
	trueLiteral  = []byte("true")
	falseLiteral = []byte("false")
	nullLiteral  = []byte("null")
)

const (
	StructLiteral    = "struct"
	InterfaceLiteral = "interface{}"
	MapLiteral       = "map"
)

type ValueType int

const (
	NotExist ValueType = iota
	String
	Number
	Float
	Object
	Array
	Boolean
	Null
	Unknown
)

func (v ValueType) String() string {
	switch v {
	case NotExist:
		return "not-exist"
	case String:
		return "string"
	case Number:
		return "number"
	case Object:
		return "object"
	case Array:
		return "array"
	case Boolean:
		return "boolean"
	case Null:
		return "null"
	case Unknown:
		return "unknown"
	default:
		return string(v)
	}
}

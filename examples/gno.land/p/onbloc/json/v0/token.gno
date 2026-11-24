package json

const (
	bracketOpen    = '['
	bracketClose   = ']'
	parenOpen      = '('
	parenClose     = ')'
	curlyOpen      = '{'
	curlyClose     = '}'
	comma          = ','
	dot            = '.'
	colon          = ':'
	backTick       = '`'
	singleQuote    = '\''
	doubleQuote    = '"'
	emptyString    = ""
	whiteSpace     = ' '
	plus           = '+'
	minus          = '-'
	aesterisk      = '*'
	bang           = '!'
	question       = '?'
	newLine        = '\n'
	tab            = '\t'
	carriageReturn = '\r'
	formFeed       = '\f'
	backSpace      = '\b'
	slash          = '/'
	backSlash      = '\\'
	underScore     = '_'
	dollarSign     = '$'
	atSign         = '@'
	andSign        = '&'
	orSign         = '|'
)

var (
	trueLiteral  = []byte("true")
	falseLiteral = []byte("false")
	nullLiteral  = []byte("null")
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
	default:
		return "unknown"
	}
}

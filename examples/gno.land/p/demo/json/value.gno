package json

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

func (v ValueType) ToString() string {
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
	}
}

var (
	TrueLiteral  = []byte("true")
	FalseLiteral = []byte("false")
	NullLiteral  = []byte("null")
)

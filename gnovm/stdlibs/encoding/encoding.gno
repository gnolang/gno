package encoding

type BinaryMarshaler interface {
	MarshalBinary() ([]byte, error)
}

type BinaryUnmarshaler interface {
	UnmarshalBinary(data []byte) error
}

type TextMarshaler interface {
	MarshalText() (text []byte, err error)
}

type TextUnmarshaler interface {
	UnmarshalText(text []byte) error
}
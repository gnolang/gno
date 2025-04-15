package types

import (
	"bytes"
	"crypto/sha256"
)

type (
	String    string
	ByteSlice []byte
)

func (self *String) MarshalBinary() ([]byte, error) {
	return []byte(*self), nil
}

func (self *String) UnmarshalBinary(data []byte) error {
	*self = String(data)
	return nil
}

func (self String) Equals(other Equatable) bool {
	if o, ok := other.(String); ok {
		return self == o
	} else {
		return false
	}
}

func (self String) Less(other Sortable) bool {
	if o, ok := other.(String); ok {
		return self < o
	} else {
		return false
	}
}

func (self String) Hash() int {
	return int(hash([]byte(self)))
}

func (self *ByteSlice) MarshalBinary() ([]byte, error) {
	return []byte(*self), nil
}

func (self *ByteSlice) UnmarshalBinary(data []byte) error {
	*self = ByteSlice(data)
	return nil
}

func (self ByteSlice) Equals(other Equatable) bool {
	if o, ok := other.(ByteSlice); ok {
		return bytes.Equal(self, o)
	} else {
		return false
	}
}

func (self ByteSlice) Less(other Sortable) bool {
	if o, ok := other.(ByteSlice); ok {
		return bytes.Compare(self, o) < 0 // -1 if a < b
	} else {
		return false
	}
}

func (self ByteSlice) Hash() int {
	return int(hash([]byte(self)))
}

func hash(s []byte) int {
	res := sha256.Sum256(s)
	return int(res[0]) |
		int(res[1]<<8) |
		int(res[2]<<16) |
		int(res[3]<<24)
}

package params

import "fmt"

// PrefixKeyMapper is used to map one key string to another.
type PrefixKeyMapper struct {
	keyMap map[string]string
}

func NewPrefixKeyMapper() PrefixKeyMapper {
	return PrefixKeyMapper{
		keyMap: map[string]string{},
	}
}

func (pkm PrefixKeyMapper) RegisterPrefix(prefix string) {
	pkm.keyMap[prefix] = "/" + prefix + "/"
}

func (pkm PrefixKeyMapper) IsExist(prefix string) bool {
	_, ok := pkm.keyMap[prefix]
	return ok
}

// Map does a transformation on an input key to produce the key
// appropriate for accessing a param keeper's storage instance.
func (pkm PrefixKeyMapper) Map(prefix string) (string, error) {
	v, ok := pkm.keyMap[prefix]
	if !ok {
		return "", fmt.Errorf("prefix %s does not exisit", prefix)
	}
	return v, nil
}

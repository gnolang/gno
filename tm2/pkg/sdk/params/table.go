package params

// This file closely mirrors the original implementation from the Cosmos SDK, with only minor modifications.

import (
	"reflect"
	"regexp"
)

type attribute struct {
	ty  reflect.Type
	vfn ValueValidatorFn
}

// KeyTable subspaces appropriate type for each parameter key
type KeyTable struct {
	m map[string]attribute
}

func NewKeyTable(pairs ...ParamSetPair) KeyTable {
	keyTable := KeyTable{
		m: make(map[string]attribute),
	}

	for _, psp := range pairs {
		keyTable = keyTable.RegisterType(psp)
	}

	return keyTable
}

// RegisterType registers a single ParamSetPair (key-type pair) in a KeyTable.
func (t KeyTable) RegisterType(psp ParamSetPair) KeyTable {
	if len(psp.Key) == 0 {
		panic("cannot register ParamSetPair with an parameter empty key")
	}
	// XXX: sanitize more?
	/*if !isAlphaNumeric(psp.Key) {
		panic("cannot register ParamSetPair with a non-alphanumeric parameter key")
	}*/
	if psp.ValidatorFn == nil {
		panic("cannot register ParamSetPair without a value validation function")
	}

	if _, ok := t.m[psp.Key]; ok {
		panic("duplicate parameter key")
	}

	rty := reflect.TypeOf(psp.Value)

	// indirect rty if it is a pointer
	for rty.Kind() == reflect.Ptr {
		rty = rty.Elem()
	}

	t.m[psp.Key] = attribute{
		vfn: psp.ValidatorFn,
		ty:  rty,
	}

	return t
}

// RegisterParamSet registers multiple ParamSetPairs from a ParamSet in a KeyTable.
func (t KeyTable) RegisterParamSet(ps ParamSet) KeyTable {
	for _, psp := range ps.ParamSetPairs() {
		t = t.RegisterType(psp)
	}
	return t
}

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

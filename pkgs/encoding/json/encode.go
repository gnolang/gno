package json

import (
	"bytes"
	"strconv"
	"sync"

	gno "github.com/gnolang/gno/pkgs/gnolang"
)

// Marshal returns the JSON encoding of v.
func Marshal(v gno.TypedValue, s gno.Store) ([]byte, error) {
	e := newEncodeState()
	defer encodeStatePool.Put(e)

	err := e.marshal(v, s, encOpts{escapeHTML: true})
	if err != nil {
		return nil, err
	}
	buf := append([]byte(nil), e.Bytes()...)

	return buf, nil
}

type encodeState struct {
	bytes.Buffer // accumulated output
}

var encodeStatePool sync.Pool

func newEncodeState() *encodeState {
	if v := encodeStatePool.Get(); v != nil {
		e := v.(*encodeState)
		e.Reset()
		return e
	}

	return &encodeState{}
}

// jsonError is an error wrapper type for internal use only.
type jsonError struct{ error }

func (e *encodeState) marshal(tv gno.TypedValue, s gno.Store, opts encOpts) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(jsonError); ok {
				err = je.error
			} else {
				panic(r)
			}
		}
	}()

	v := newValue(tv, s)
	valueEncoder(v)(e, v, opts)
	return nil
}

type encOpts struct {
	escapeHTML bool
	quoted     bool
}

type encoderFunc func(e *encodeState, v Value, opts encOpts)

func valueEncoder(v Value) encoderFunc {
	switch v.Kind() {
	case gno.InvalidKind:
		return invalidValueEncoder
	case gno.BoolKind:
		return boolEncoder
	case gno.IntKind, gno.Int8Kind, gno.Int16Kind, gno.Int32Kind, gno.Int64Kind:
		return intEncoder
	case gno.UintKind, gno.Uint8Kind, gno.Uint16Kind, gno.Uint32Kind, gno.Uint64Kind:
		return uintEncoder
	case gno.StringKind:
		return stringEncoder
	case gno.ArrayKind:
		return arrayEncoder
	case gno.SliceKind:
		return sliceEncoder
	case gno.StructKind:
		return structEncoder
	case gno.PointerKind:
		return ptrEncoder
	case gno.InterfaceKind:
		return interfaceEncoder
	default:
		panic("unreachable") //todo: unsupportedTypeEncoder?
	}
}

func invalidValueEncoder(e *encodeState, v Value, _ encOpts) {
	e.WriteString("null")
}

func boolEncoder(e *encodeState, v Value, opts encOpts) {
	if opts.quoted {
		e.WriteByte('"')
	}
	if v.Bool() {
		e.WriteString("true")
	} else {
		e.WriteString("false")
	}
	if opts.quoted {
		e.WriteByte('"')
	}
}

func stringEncoder(e *encodeState, v Value, opts encOpts) {
	e.WriteByte('"')
	e.WriteString(v.String())
	e.WriteByte('"')
}

func intEncoder(e *encodeState, v Value, opts encOpts) {
	if opts.quoted {
		e.WriteByte('"')
	}

	e.WriteString(strconv.FormatInt(v.Int(), 10))

	if opts.quoted {
		e.WriteByte('"')
	}
}

func uintEncoder(e *encodeState, v Value, opts encOpts) {
	if opts.quoted {
		e.WriteByte('"')
	}

	e.WriteString(strconv.FormatUint(v.Uint(), 10))

	if opts.quoted {
		e.WriteByte('"')
	}
}

func arrayEncoder(e *encodeState, v Value, opts encOpts) {
	e.WriteByte('[')

	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			e.WriteByte(',')
		}
		elem := v.Index(i)
		valueEncoder(elem)(e, elem, opts)
	}

	e.WriteByte(']')
}

func sliceEncoder(e *encodeState, v Value, opts encOpts) {
	e.WriteByte('[')

	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			e.WriteByte(',')
		}
		elem := v.Index(i)
		valueEncoder(elem)(e, elem, opts)
	}

	e.WriteByte(']')
}

func structEncoder(e *encodeState, v Value, opts encOpts) {
	next := byte('{')

	fields := v.StructFields()

	for i := 0; i < len(fields); i++ {
		field := fields[i]
		if field.IsZero() {
			continue
		}
		e.WriteByte(next)
		next = ','

		e.WriteByte('"')
		e.WriteString(field.Name())
		e.WriteByte('"')
		e.WriteByte(':')

		valueEncoder(field.Value())(e, field.Value(), opts)
	}

	if next == '{' {
		e.WriteString("{}")
	} else {
		e.WriteByte('}')
	}
}

func ptrEncoder(e *encodeState, v Value, opts encOpts) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	valueEncoder(v.Elem())(e, v.Elem(), opts)
}

func interfaceEncoder(e *encodeState, v Value, opts encOpts) {
	if v.IsNil() {
		e.WriteString("null")
		return
	}
	valueEncoder(v.Elem())(e, v.Elem(), opts)
}

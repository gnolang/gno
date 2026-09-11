package toml

import (
	"fmt"
	"reflect"
	"time"
)

var kindToType = [reflect.String + 1]reflect.Type{
	reflect.Bool:    reflect.TypeFor[bool](),
	reflect.String:  reflect.TypeFor[string](),
	reflect.Float32: reflect.TypeFor[float64](),
	reflect.Float64: reflect.TypeFor[float64](),
	reflect.Int:     reflect.TypeFor[int64](),
	reflect.Int8:    reflect.TypeFor[int64](),
	reflect.Int16:   reflect.TypeFor[int64](),
	reflect.Int32:   reflect.TypeFor[int64](),
	reflect.Int64:   reflect.TypeFor[int64](),
	reflect.Uint:    reflect.TypeFor[uint64](),
	reflect.Uint8:   reflect.TypeFor[uint64](),
	reflect.Uint16:  reflect.TypeFor[uint64](),
	reflect.Uint32:  reflect.TypeFor[uint64](),
	reflect.Uint64:  reflect.TypeFor[uint64](),
}

// typeFor returns a reflect.Type for a reflect.Kind, or nil if none is found.
// supported values:
// string, bool, int64, uint64, float64, time.Time, int, int8, int16, int32, uint, uint8, uint16, uint32, float32
func typeFor(k reflect.Kind) reflect.Type {
	if k > 0 && int(k) < len(kindToType) {
		return kindToType[k]
	}
	return nil
}

func simpleValueCoercion(object any) (any, error) {
	switch original := object.(type) {
	case string, bool, int64, uint64, float64, time.Time:
		return original, nil
	case int:
		return int64(original), nil
	case int8:
		return int64(original), nil
	case int16:
		return int64(original), nil
	case int32:
		return int64(original), nil
	case uint:
		return uint64(original), nil
	case uint8:
		return uint64(original), nil
	case uint16:
		return uint64(original), nil
	case uint32:
		return uint64(original), nil
	case float32:
		return float64(original), nil
	case fmt.Stringer:
		return original.String(), nil
	case []any:
		value := reflect.ValueOf(original)
		length := value.Len()
		arrayValue := reflect.MakeSlice(value.Type(), 0, length)
		for i := range length {
			val := value.Index(i).Interface()
			simpleValue, err := simpleValueCoercion(val)
			if err != nil {
				return nil, err
			}
			arrayValue = reflect.Append(arrayValue, reflect.ValueOf(simpleValue))
		}
		return arrayValue.Interface(), nil
	default:
		return nil, fmt.Errorf("cannot convert type %T to Tree", object)
	}
}

func sliceToTree(object any) (any, error) {
	// arrays are a bit tricky, since they can represent either a
	// collection of simple values, which is represented by one
	// *tomlValue, or an array of tables, which is represented by an
	// array of *Tree.

	// holding the assumption that this function is called from toTree only when value.Kind() is Array or Slice
	value := reflect.ValueOf(object)
	insideType := value.Type().Elem()
	length := value.Len()
	if length > 0 {
		insideType = reflect.ValueOf(value.Index(0).Interface()).Type()
	}
	if insideType.Kind() == reflect.Map {
		// this is considered as an array of tables
		tablesArray := make([]*Tree, 0, length)
		for i := range length {
			table := value.Index(i)
			tree, err := toTree(table.Interface())
			if err != nil {
				return nil, err
			}
			tablesArray = append(tablesArray, tree.(*Tree))
		}
		return tablesArray, nil
	}

	sliceType := typeFor(insideType.Kind())
	if sliceType == nil {
		sliceType = insideType
	}

	arrayValue := reflect.MakeSlice(reflect.SliceOf(sliceType), 0, length)

	for i := range length {
		val := value.Index(i).Interface()
		simpleValue, err := simpleValueCoercion(val)
		if err != nil {
			return nil, err
		}
		arrayValue = reflect.Append(arrayValue, reflect.ValueOf(simpleValue))
	}
	return &tomlValue{value: arrayValue.Interface(), position: Position{}}, nil
}

func toTree(object any) (any, error) {
	value := reflect.ValueOf(object)

	if value.Kind() == reflect.Map {
		values := map[string]any{}
		keys := value.MapKeys()
		for _, key := range keys {
			if key.Kind() != reflect.String {
				if _, ok := key.Interface().(string); !ok {
					return nil, fmt.Errorf("map key needs to be a string, not %T (%v)", key.Interface(), key.Kind())
				}
			}

			v := value.MapIndex(key)
			newValue, err := toTree(v.Interface())
			if err != nil {
				return nil, err
			}
			values[key.String()] = newValue
		}
		return &Tree{values: values, position: Position{}}, nil
	}

	if value.Kind() == reflect.Array || value.Kind() == reflect.Slice {
		return sliceToTree(object)
	}

	simpleValue, err := simpleValueCoercion(object)
	if err != nil {
		return nil, err
	}
	return &tomlValue{value: simpleValue, position: Position{}}, nil
}

package vm

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

func UnmarshalJSON(alloc *gno.Allocator, store gno.Store, b []byte, t gno.Type) (*gno.TypedValue, error) {
	var tv gno.TypedValue
	tv.T = t
	gvalue := gnolang.Gno2GoValue(&tv, reflect.Value{})
	v := reflect.New(gvalue.Type())
	err := amino.UnmarshalJSON(b, v.Interface())
	if err != nil {
		return nil, err
	}

	typedvalue := gnolang.Go2GnoValue(alloc, store, v.Elem())
	return &typedvalue, nil
}

func UnmarshalNativeValueJSON(b []byte, t gno.Type) (*gno.TypedValue, error) {
	var tv gno.TypedValue
	tv.T = t
	gvalue := gnolang.Gno2GoValue(&tv, reflect.Value{})
	v := reflect.New(gvalue.Type())
	err := amino.UnmarshalJSON(b, v.Interface())
	if err != nil {
		return nil, err
	}

	typedvalue := gnolang.Go2GnoNativeValue(nil, v.Elem())
	return &typedvalue, nil
}

func MarshalJSON(tv *gno.TypedValue) ([]byte, error) {
	fmt.Println("ret tv:", tv)
	rv := gnolang.Gno2GoValue(tv, reflect.Value{})
	return json.Marshal(rv.Interface())
}

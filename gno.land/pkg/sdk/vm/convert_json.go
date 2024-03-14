package vm

import (
	"encoding/json"
	"reflect"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

func UnmarshalJSON(alloc *gno.Allocator, store gno.Store, b []byte, t gno.Type) (tv gno.TypedValue, err error) {
	tv.T = t
	gvalue := gnolang.Gno2GoValue(&tv, reflect.Value{})
	v := reflect.New(gvalue.Type())
	if err := amino.UnmarshalJSON(b, v.Interface()); err != nil {
		return tv, err
	}

	return gnolang.Go2GnoValue(alloc, store, v.Elem()), nil
}

func UnmarshalNativeValueJSON(alloc *gno.Allocator, b []byte, t gno.Type) (tv gno.TypedValue, err error) {
	tv.T = t
	gvalue := gnolang.Gno2GoValue(&tv, reflect.Value{})
	v := reflect.New(gvalue.Type())
	if err = amino.UnmarshalJSON(b, v.Interface()); err != nil {
		return tv, err
	}

	return gnolang.Go2GnoNativeValue(alloc, v.Elem()), nil
}

func MarshalJSON(tv *gno.TypedValue) ([]byte, error) {
	rv := gnolang.Gno2GoValue(tv, reflect.Value{})
	// XXX: use amino
	return json.Marshal(rv.Interface())
}

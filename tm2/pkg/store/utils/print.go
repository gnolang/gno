package utils

import (
	"fmt"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/colors"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

type Printer interface {
	Print()
}

// TODO move to another file.
func Print(store types.Store) {
	fmt.Println(colors.Blue("//----------------------------------------"))
	if ps, ok := store.(Printer); ok {
		ps.Print()
	} else {
		fmt.Println(colors.Blue(fmt.Sprintf("// store:%p %v", store, reflect.TypeOf(store))))
		itr := store.Iterator(nil, nil)
		defer itr.Close()
		for ; itr.Valid(); itr.Next() {
			key, value := itr.Key(), itr.Value()
			var keystr, valuestr string
			keystr = string(colors.ColoredBytes(key, colors.Green, colors.Blue))
			valuestr = fmt.Sprintf("(%d)", len(value))
			/*
				if true || strings.IsASCIIText(string(value)) {
					valuestr = string(value)
				} else {
					valuestr = fmt.Sprintf("0x%X", value)
				}
			*/
			fmt.Printf("%s: %s\n", keystr, valuestr)
		}
	}
	fmt.Println(colors.Blue("//------------------------------------ end"))
}

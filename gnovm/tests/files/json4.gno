package main

import (
    "fmt"
    
	"gno.land/p/demo/json"
)

// dynamic struct creation and marshal to json
func main() {
    s := json.NewStruct()

    for i:=0; i < 10; i++ {
        s.AddIntField(fmt.Sprintf("field%d", i), i)	// <- add field in runtime
    }

    str, err := s.Marshal()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))
}
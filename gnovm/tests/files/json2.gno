package main

import (
	"fmt"

	"gno.land/p/demo/json"
)

func main() {
	data := []byte(`{"name": "Foo", "age": 20, "score": 4.2, "subscribed": true, "option": null}`)

	s := json.NewStruct()
	err := json.Unmarshal(data, s)
	if err != nil {
		panic(err)
	}

	fmt.Println(s.String())
}
package main

import (
	"fmt"

	"gno.land/p/demo/json"
)

// Unmarshal JSON data into a struct.
func main() {
	data := []byte(`{"name": "Foo", "age": 20, "score": 4.2, "subscribed": true, "option": null}`)

	s := json.NewStruct()
	err := json.Unmarshal(data, s)
	if err != nil {
		panic(err)
	}

	// `String()` method use Marshal internally to generate human readable JSON.
	fmt.Println(s.String())
}

// Output:
// {"name": "Foo", "age": 20, "score": 4.2, "subscribed": true, "option": null}
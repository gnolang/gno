package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/json"
)


func main() {
	data := []byte(`{"Name": "John Doe", "Age": 30, "IsActive": true}`)
	s := json.NewStruct().
		AddStringField("Name", "John Doe").
		AddIntField("Age", 30).
		AddBoolField("IsActive", true).
		AddFloatField("Score", 4.5)

	jsonStr, err := s.Marshal() 
	if err != nil {
		fmt.Println("Error marshaling:", err)
		return
	}
	println("Marshalled:")
	println(jsonStr)
	println("--------------------")

	if json.Unmarshal(data, s); err != nil {
		fmt.Println("Error unmarshaling:", err)
		return
	}

	println("Unmarshalled:")
	for _, field := range s.Fields {
        fmt.Printf("%s: %v\n", field.Name, field.Value)
    }
}


// output:
// Marshalled:
// {"Name": "John Doe", "Age": 30, "IsActive": true, "Score": 4.5}
// --------------------
// Unmarshalled:
// Name: John Doe
// Age: 30
// IsActive: true
// Score: 4.5
package main

import (
	"fmt"

	"gno.land/p/demo/json"
)


func main() {
	data := []byte(`{"Name": "John Doe", "Age": 30, "IsActive": true}`)
	s := json.NewStruct().
		AddStringField("Name", "John Doe").
		AddIntField("Age", 30).
		AddBoolField("IsActive", true).
		AddFloatField("Score", 4.5).
		AddNullField("Option")

	jsonStr, err := s.Marshaler() 
	if err != nil {
		fmt.Println("Error marshaling:", err)
		return
	}
	println("Marshalled:")
	println(string(jsonStr))
	println("--------------------")

	if json.Unmarshaler(data, s); err != nil {
		fmt.Println("Error unmarshaling:", err)
		return
	}

	println("Unmarshalled:")
	for _, field := range s.Fields {
        fmt.Printf("%s: %v\n", field.Name, field.Value)
    }
}


// Output:
// Marshalled:
// {"Name": "John Doe", "Age": 30, "IsActive": true, "Score": 4.5, "Option": null}
// --------------------
// Unmarshalled:
// Name: John Doe
// Age: 30
// IsActive: true
// Score: 4.5
// Option: <nil>
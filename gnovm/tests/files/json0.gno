package main

import (
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/json"
)


func main() {
	s := json.NewStruct().
		AddStringField("Name", "John Doe").
		AddIntField("Age", 30).
		AddBoolField("IsActive", true).
		AddFloatField("Score", 4.5).
		AddNullField("Option").
		AddArrayField("arr", []interface{}{"one", 2, true, nil})

	jsonStr, err := s.Marshal() 
	if err != nil {
		ufmt.Println("Error marshaling:", err)
		return
	}
	ufmt.Println("Marshalled:")
	ufmt.Println(string(jsonStr))
}

// Output:
// Marshalled:
// {"Name": "John Doe", "Age": 30, "IsActive": true, "Score": 4.5, "Option": null, "arr": ["one", 2, true, null]}
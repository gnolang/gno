package main

import (
	"fmt"

	"gno.land/p/demo/json"
)

func main() {
    contactStruct := json.NewStruct().
        AddStringField("email", "john@example.com").
        AddStringField("phone", "123-456-7890")

    addressStruct := json.NewStruct().
        AddStringField("street", "123 Main St").
        AddStringField("city", "Anytown").
        AddStringField("state", "CA").
        AddIntField("zip", 12345)

    userStruct := json.NewStruct().
        AddStringField("name", "John Doe").
        AddObjectField("contact", contactStruct).
        AddObjectField("address", addressStruct).
        AddArrayField("roles", []interface{}{"admin", "user"})

    s := json.NewStruct().
        AddObjectField("user", userStruct)

    jsonStr, err := s.Marshaler()
    if err != nil {
        fmt.Println("Error marshaling:", err)
        return
    }

    fmt.Println(string(jsonStr))
}

// Ouput:
// {"user": {"name": "John Doe", "contact": {"email": "john@example.com", "phone": "123-456-7890"}, "address": {"street": "123 Main St", "city": "Anytown", "state": "CA", "zip": 12345}, "roles": ["admin", "user"]}}
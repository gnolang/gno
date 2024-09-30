package main

import "fmt"

type Role struct {
	Name        string
	Permissions []string
	Users       []string
	Next        *Role
	Prev        *Role
}

func main() {
	userRole := &Role{
		Name:        "user",
		Permissions: []string{},
		Users:       []string{},
		Next:        nil,
		Prev:        nil,
	}

	fmt.Printf("%v", userRole)
}

// Output:
// &{user [] [] <nil> <nil>}

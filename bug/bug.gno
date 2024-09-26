package main

import "fmt"

type Role struct {
	name        string
	permissions []string
	users       []string 
	next        *Role 
	prev        *Role 
}

func main() {
	userRole := &Role{
		name:        "user",
		permissions: []string{},
		users:       []string{},
		next:        nil,
		prev:        nil,
	}

	fmt.Printf("%v", userRole)
}

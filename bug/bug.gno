package main

import "fmt"

type Role struct {
	Name        string
	Permissions []string
	Users       []string
	Next        *Role
	Prev        *Role
	PermissionData *Permission
}

type Permission struct {
 Name string
 Resource string
}

func main() {
	userRole := &Role{
		Name:        "user",
		Permissions: []string{},
		Users:       []string{},
		Next:        nil,
		Prev:        nil,
		PermissionData: &Permission{"Hello", "Test"},
	}

	fmt.Printf("%v", userRole)
}

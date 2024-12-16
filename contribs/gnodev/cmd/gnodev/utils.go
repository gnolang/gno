package main

import "strings"

type varStrings []string

func (va *varStrings) Set(val string) error {
	for _, subval := range strings.Split(val, ",") {
		*va = append(*va, subval)
	}

	return nil
}

func (va *varStrings) String() string {
	return strings.Join(*va, ",")
}

func (va *varStrings) Strings() []string {
	if va == nil {
		return []string{}
	}

	return []string(*va)
}

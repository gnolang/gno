package main

import (
	"encoding/json"
	"fmt"
)

type Email struct {
	Where string
	Addr  string
}

func f(s string, r interface{}) interface{} {
	return json.Unmarshal([]byte(s), &r)
}

func main() {
	data := `
		{"Where": "work", "Addr": "bob@work.com"}
	`
	v := Email{}
	err := f(data, &v)
	fmt.Println(err, v)
}

// Output:
// <nil> {work bob@work.com}

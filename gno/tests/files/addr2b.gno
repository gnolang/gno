package main

import (
	"encoding/xml"
	"fmt"
)

type Email struct {
	Where string `xml:"where,attr"`
	Addr  string
}

func f(s string, r interface{}) interface{} {
	return xml.Unmarshal([]byte(s), &r)
}

func main() {
	data := `
		<Email where='work'>
			<Addr>bob@work.com</Addr>
		</Email>
	`
	v := Email{}
	err := f(data, &v)
	fmt.Println(err, v)
}

// Output:
// <nil> {work bob@work.com}

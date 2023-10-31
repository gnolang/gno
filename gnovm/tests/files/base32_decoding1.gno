package main

import (
	"encoding/base32"
	"fmt"
)

func main() {
	str := "ONXW2ZJAMRQXIYJAO5UXI2BAAAQGC3TEEDX3XPY="
	data, err := base32.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%q\n", data)
}

// Output:
// "some data with \x00 and \ufeff"
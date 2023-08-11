package gnoclient_test

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
)

func Example() {
	client := gnoclient.Client{}
	_ = client
	fmt.Println("Hello")
	// Output:
	// Hello
}

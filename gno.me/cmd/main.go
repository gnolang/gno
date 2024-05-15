package main

import (
	"fmt"

	"github.com/gnolang/gno/gno.me/examples"
	"github.com/gnolang/gno/gno.me/gno"
	"github.com/gnolang/gno/gno.me/http"
)

func main() {
	vm := gno.NewVM()
	fmt.Println("VM created")
	if err := examples.CreateInstallerApp(vm); err != nil {
		panic("could not add installer realm: " + err.Error())
	}

	server := http.NewServer(vm)
	fmt.Println("Starting server...")
	server.ListenAndServe()
}

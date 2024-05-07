package http

import (
	gohttp "net/http"

	"github.com/gnolang/gno/gno.me/gno"
)

var vm gno.VM

func NewServer(gnoVM gno.VM) gohttp.Server {
	vm = gnoVM

	mux := gohttp.NewServeMux()
	mux.HandleFunc("/system/install", installApp)
	mux.HandleFunc("/system/call", callApp)
	mux.HandleFunc("/system/run", run)
	mux.HandleFunc("/", renderApp)

	return gohttp.Server{
		Addr:    ":4591",
		Handler: mux,
	}
}

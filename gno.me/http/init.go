package http

import (
	"context"
	gohttp "net/http"

	"github.com/gnolang/gno/gno.me/gno"
	"github.com/gnolang/gno/gno.me/ws"
)

var (
	vm                   gno.VM
	eventListenerManager *ws.Manager
	wsPort               string
)

func newMux() *gohttp.ServeMux {
	mux := gohttp.NewServeMux()
	mux.HandleFunc("/system/create", createApp)
	mux.HandleFunc("/system/call", callApp)
	mux.HandleFunc("/system/run", run)
	mux.HandleFunc("/", renderApp)
	return mux
}

func NewServer(gnoVM gno.VM, port string) *gohttp.Server {
	vm = gnoVM
	mux := newMux()

	// Overwrite the existing port number on each startup.
	if _, _, err := vm.Call(context.Background(), "port", false, "Set", port); err != nil {
		panic("error setting port: " + err.Error())
	}

	return &gohttp.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
}

func NewServerWithRemoteSupport(gnoVM gno.VM, manager *ws.Manager, httpPort string, socketPort string) *gohttp.Server {
	vm = gnoVM
	eventListenerManager = manager
	wsPort = socketPort

	mux := newMux()
	mux.HandleFunc("/system/install-remote", installRemoteApp)
	mux.HandleFunc("/system/get-app", getApp)

	return &gohttp.Server{
		Addr:    ":" + httpPort,
		Handler: mux,
	}
}

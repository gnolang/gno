package http

import (
	gohttp "net/http"

	"github.com/gnolang/gno/gno.me/gno"
	"github.com/gnolang/gno/gno.me/ws"
)

var (
	vm                   gno.VM
	eventListenerManager *ws.Manager
)

func newMux() *gohttp.ServeMux {
	mux := gohttp.NewServeMux()
	mux.HandleFunc("/system/create", createApp)
	mux.HandleFunc("/system/call", callApp)
	mux.HandleFunc("/system/run", run)
	mux.HandleFunc("/", renderApp)
	return mux
}

func NewServer(gnoVM gno.VM) *gohttp.Server {
	vm = gnoVM
	mux := newMux()

	return &gohttp.Server{
		Addr:    ":4591",
		Handler: mux,
	}
}

func NewServerWithRemoteSupport(gnoVM gno.VM, manager *ws.Manager) *gohttp.Server {
	vm = gnoVM
	eventListenerManager = manager
	mux := newMux()
	mux.HandleFunc("/system/install-remote", installRemoteApp)
	mux.HandleFunc("/system/get-app", getApp)

	return &gohttp.Server{
		Addr:    ":4591",
		Handler: mux,
	}
}

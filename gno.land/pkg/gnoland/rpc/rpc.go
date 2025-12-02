package rpc

import (
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	rpcserver "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
)

type Server struct {
	app    *gnoland.App
	logger *slog.Logger
}

func NewServer(app *gnoland.App, logger *slog.Logger) *Server {
	return &Server{
		app:    app,
		logger: logger.With("module", "gnoland_rpc"),
	}
}

func (s *Server) rpcFuncs() map[string]*rpcserver.RPCFunc {
	return map[string]*rpcserver.RPCFunc{
		"vm/render":  rpcserver.NewRPCFunc(s.VMRender, "height,pkgPath,path"),
		"vm/funcs":   rpcserver.NewRPCFunc(s.VMFuncs, "height,pkgPath"),
		"vm/eval":    rpcserver.NewRPCFunc(s.VMEval, "height,data"),
		"vm/file":    rpcserver.NewRPCFunc(s.VMFile, "height,filepath"),
		"vm/doc":     rpcserver.NewRPCFunc(s.VMDoc, "height,pkgPath"),
		"vm/paths":   rpcserver.NewRPCFunc(s.VMPaths, "height,target,limit"),
		"vm/storage": rpcserver.NewRPCFunc(s.VMStorage, "height,pkgPath"),
	}
}

func (s *Server) NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register the HTTP handlers
	rpcserver.RegisterRPCFuncs(mux, s.rpcFuncs(), s.logger)

	// Register the websocket handlers as well
	wsMgr := rpcserver.NewWebsocketManager(s.rpcFuncs())
	wsMgr.SetLogger(s.logger)
	mux.HandleFunc("/websocket", wsMgr.WebsocketHandler)

	return mux
}

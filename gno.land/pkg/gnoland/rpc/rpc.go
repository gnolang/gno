package rpc

import (
	"log/slog"
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	rpcserver "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// Application is the required Gnoland app abstraction for the RPC
type Application interface {
	// NewQueryContext creates a new app query context (read-only)
	NewQueryContext(height int64) (sdk.Context, error)

	// VMKeeper returns the VM keeper associated with the app
	VMKeeper() vm.VMKeeperI
}

// Server is the Gnoland (app) RPC server instance
type Server struct {
	app    Application
	logger *slog.Logger
}

// NewServer creates a new instance of the Gnoland (app) RPC server
func NewServer(app Application, logger *slog.Logger) *Server {
	return &Server{
		app:    app,
		logger: logger.With("module", "gnoland_rpc"),
	}
}

// rpcFuncs returns the endpoint -> handler mapping
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

// NewMux creates a server mux, and registers the endpoints for both http and ws requests
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

package main

import (
	"fmt"
	"net/http"
	"os"

	amino "github.com/tendermint/go-amino-x"

	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/classic/libs/log"
	rpcserver "github.com/tendermint/classic/rpc/lib/server"
	rpctypes "github.com/tendermint/classic/rpc/lib/types"
)

var routes = map[string]*rpcserver.RPCFunc{
	"hello_world": rpcserver.NewRPCFunc(HelloWorld, "name,num"),
}

func HelloWorld(ctx *rpctypes.Context, name string, num int) (Result, error) {
	return Result{fmt.Sprintf("hi %s %d", name, num)}, nil
}

type Result struct {
	Result string
}

func main() {
	var (
		mux    = http.NewServeMux()
		cdc    = amino.NewCodec()
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	)

	// Stop upon receiving SIGTERM or CTRL-C.
	cmn.TrapSignal(logger, func() {})

	rpcserver.RegisterRPCFuncs(mux, routes, cdc, logger)
	config := rpcserver.DefaultConfig()
	listener, err := rpcserver.Listen("0.0.0.0:8008", config)
	if err != nil {
		cmn.Exit(err.Error())
	}
	rpcserver.StartHTTPServer(listener, mux, logger, config)
}

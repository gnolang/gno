package rpc

import (
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	http2 "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/http"
	client "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/uri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	server "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// Client and Server should work over tcp or unix sockets
const (
	tcpAddr                  = "tcp://0.0.0.0:47768"
	tcpServerUnavailableAddr = "tcp://0.0.0.0:47769"

	unixSocket = "/tmp/rpc_test.sock"
	unixAddr   = "unix://" + unixSocket

	websocketEndpoint = "/websocket/endpoint"
)

type ResultEcho struct {
	Value string `json:"value"`
}

type ResultEchoInt struct {
	Value int `json:"value"`
}

type ResultEchoBytes struct {
	Value []byte `json:"value"`
}

type ResultEchoDataBytes struct {
	Value []byte `json:"value"`
}

// Define some routes
var Routes = map[string]*server.RPCFunc{
	"echo":            server.NewRPCFunc(EchoResult, "arg"),
	"echo_ws":         server.NewWSRPCFunc(EchoWSResult, "arg"),
	"echo_bytes":      server.NewRPCFunc(EchoBytesResult, "arg"),
	"echo_data_bytes": server.NewRPCFunc(EchoDataBytesResult, "arg"),
	"echo_int":        server.NewRPCFunc(EchoIntResult, "arg"),
}

func EchoResult(ctx *types.Context, v string) (*ResultEcho, error) {
	return &ResultEcho{v}, nil
}

func EchoWSResult(ctx *types.Context, v string) (*ResultEcho, error) {
	return &ResultEcho{v}, nil
}

func EchoIntResult(ctx *types.Context, v int) (*ResultEchoInt, error) {
	return &ResultEchoInt{v}, nil
}

func EchoBytesResult(ctx *types.Context, v []byte) (*ResultEchoBytes, error) {
	return &ResultEchoBytes{v}, nil
}

func EchoDataBytesResult(ctx *types.Context, v []byte) (*ResultEchoDataBytes, error) {
	return &ResultEchoDataBytes{v}, nil
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

// launch unix and tcp servers
func setup() {
	logger := log.NewNoopLogger()

	cmd := exec.Command("rm", "-f", unixSocket)
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	if err = cmd.Wait(); err != nil {
		panic(err)
	}

	tcpLogger := logger.With("socket", "tcp")
	mux := http.NewServeMux()
	server.RegisterRPCFuncs(mux, Routes, tcpLogger)
	wm := server.NewWebsocketManager(Routes, server.ReadWait(5*time.Second), server.PingPeriod(1*time.Second))
	wm.SetLogger(tcpLogger)
	mux.HandleFunc(websocketEndpoint, wm.WebsocketHandler)
	config := server.DefaultConfig()
	listener1, err := server.Listen(tcpAddr, config)
	if err != nil {
		panic(err)
	}
	go server.StartHTTPServer(listener1, mux, tcpLogger, config)

	unixLogger := logger.With("socket", "unix")
	mux2 := http.NewServeMux()
	server.RegisterRPCFuncs(mux2, Routes, unixLogger)
	wm = server.NewWebsocketManager(Routes)
	wm.SetLogger(unixLogger)
	mux2.HandleFunc(websocketEndpoint, wm.WebsocketHandler)
	listener2, err := server.Listen(unixAddr, config)
	if err != nil {
		panic(err)
	}
	go server.StartHTTPServer(listener2, mux2, unixLogger, config)

	listener3, err := server.Listen(tcpServerUnavailableAddr, config)
	if err != nil {
		panic(err)
	}
	mux3 := http.NewServeMux()
	mux3.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "oups", http.StatusTeapot)
	})
	go server.StartHTTPServer(listener3, mux3, tcpLogger, config)

	// wait for servers to start
	time.Sleep(time.Second * 2)
}

func echoViaHTTP(cl http2.HTTPClient, val string) (string, error) {
	params := map[string]interface{}{
		"arg": val,
	}
	result := new(ResultEcho)
	if err := cl.Call("echo", params, result); err != nil {
		return "", err
	}
	return result.Value, nil
}

// -------------
func TestHexStringArg(t *testing.T) {
	t.Parallel()

	cl := client.NewClient(tcpAddr)
	// should NOT be handled as hex
	val := "0xabc"
	got, err := echoViaHTTP(cl, val)
	require.Nil(t, err)
	assert.Equal(t, got, val)
}

func TestQuotedStringArg(t *testing.T) {
	t.Parallel()

	cl := client.NewClient(tcpAddr)
	// should NOT be unquoted
	val := "\"abc\""
	got, err := echoViaHTTP(cl, val)
	require.Nil(t, err)
	assert.Equal(t, got, val)
}

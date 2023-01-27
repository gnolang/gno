package rpctest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	cfg "github.com/gnolang/gno/pkgs/bft/config"
	nm "github.com/gnolang/gno/pkgs/bft/node"
	"github.com/gnolang/gno/pkgs/bft/privval"
	"github.com/gnolang/gno/pkgs/bft/proxy"
	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	rpcclient "github.com/gnolang/gno/pkgs/bft/rpc/lib/client"
	"github.com/gnolang/gno/pkgs/log"
	osm "github.com/gnolang/gno/pkgs/os"
	"github.com/gnolang/gno/pkgs/p2p"
)

// Options helps with specifying some parameters for our RPC testing for greater
// control.
type Options struct {
	suppressStdout bool
	recreateConfig bool
}

var (
	globalConfig   *cfg.Config
	defaultOptions = Options{
		suppressStdout: false,
		recreateConfig: false,
	}
)

func waitForRPC() {
	laddr := GetConfig().RPC.ListenAddress
	client := rpcclient.NewJSONRPCClient(laddr)
	result := new(ctypes.ResultStatus)
	for {
		_, err := client.Call("status", map[string]interface{}{}, result)
		if err == nil {
			return
		} else {
			fmt.Println("error", err)
			time.Sleep(time.Millisecond)
		}
	}
}

// f**ing long, but unique for each test
func makePathname() string {
	// get path
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// fmt.Println(p)
	sep := string(filepath.Separator)

	return strings.Replace(p, sep, "_", -1)
}

func randPort() int {
	port, err := osm.GetFreePort()
	if err != nil {
		panic(err)
	}

	return port
}

func makeAddrs() (string, string) {
	return fmt.Sprintf("tcp://0.0.0.0:%d", randPort()),
		fmt.Sprintf("tcp://0.0.0.0:%d", randPort())
}

func createConfig() *cfg.Config {
	pathname := makePathname()
	c := cfg.ResetTestRoot(pathname)

	// and we use random ports to run in parallel
	tm, rpc := makeAddrs()
	c.P2P.ListenAddress = tm
	c.RPC.ListenAddress = rpc
	c.RPC.CORSAllowedOrigins = []string{"https://tendermint.com/"}
	// c.TxIndex.IndexTags = "app.creator,tx.height" // see kvstore application

	return c
}

// GetConfig returns a config for the test cases as a singleton
func GetConfig(forceCreate ...bool) *cfg.Config {
	if globalConfig == nil || (len(forceCreate) > 0 && forceCreate[0]) {
		globalConfig = createConfig()
	}

	return globalConfig
}

// StartTendermint starts a test tendermint server in a go routine and returns when it is initialized
func StartTendermint(app abci.Application, opts ...func(*Options)) *nm.Node {
	nodeOpts := defaultOptions
	for _, opt := range opts {
		opt(&nodeOpts)
	}
	node := NewTendermint(app, &nodeOpts)
	err := node.Start()
	if err != nil {
		panic(err)
	}

	// wait for rpc
	waitForRPC()

	if !nodeOpts.suppressStdout {
		fmt.Println("Tendermint running!")
	}

	return node
}

// StopTendermint stops a test tendermint server, waits until it's stopped and
// cleans up test/config files.
func StopTendermint(node *nm.Node) {
	node.Stop()
	node.Wait()
	os.RemoveAll(node.Config().RootDir)
}

// NewTendermint creates a new tendermint server and sleeps forever
func NewTendermint(app abci.Application, opts *Options) *nm.Node {
	// Create & start node
	config := GetConfig(opts.recreateConfig)
	var logger log.Logger
	if opts.suppressStdout {
		logger = log.NewNopLogger()
	} else {
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		logger.SetLevel(log.LevelError)
	}
	pvKeyFile := config.PrivValidatorKeyFile()
	pvKeyStateFile := config.PrivValidatorStateFile()
	pv := privval.LoadOrGenFilePV(pvKeyFile, pvKeyStateFile)
	papp := proxy.NewLocalClientCreator(app)
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		panic(err)
	}
	node, err := nm.NewNode(config, pv, nodeKey, papp,
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		logger)
	if err != nil {
		panic(err)
	}

	return node
}

// SuppressStdout is an option that tries to make sure the RPC test Tendermint
// node doesn't log anything to stdout.
func SuppressStdout(o *Options) {
	o.suppressStdout = true
}

// RecreateConfig instructs the RPC test to recreate the configuration each
// time, instead of treating it as a global singleton.
func RecreateConfig(o *Options) {
	o.recreateConfig = true
}

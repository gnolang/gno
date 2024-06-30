package client

import (
	"context"
	"flag"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

type QueryCfg struct {
	RootCfg *BaseCfg

	Data string
	Path string
}

const balancesQuery = "bank/balances/"

func NewQueryCmd(rootCfg *BaseCfg, io commands.IO) *commands.Command {
	cfg := &QueryCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "query",
			ShortUsage: "query [flags] <path>",
			ShortHelp:  "makes an ABCI query",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execQuery(cfg, args, io)
		},
	)
}

func (c *QueryCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.Data,
		"data",
		"",
		"query data bytes",
	)
}

func execQuery(cfg *QueryCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	cfg.Path = args[0]

	qres, err := QueryHandler(cfg)
	if err != nil {
		return err
	}

	if qres.Response.Error != nil {
		io.Printf("Log: %s\n",
			qres.Response.Log)
		return qres.Response.Error
	}

	resdata := qres.Response.Data
	// XXX in general, how do we know what to show?
	// proof := qres.Response.Proof
	height := qres.Response.Height
	io.Printf("height: %d\ndata: %s\n",
		height,
		string(resdata))
	return nil
}

func QueryHandler(cfg *QueryCfg) (*ctypes.ResultABCIQuery, error) {
	remote := cfg.RootCfg.Remote
	if remote == "" {
		return nil, errors.New("missing remote url")
	}

	path := cfg.Path

	if strings.Contains(path, balancesQuery) {
		path = handleBalanceQuery(path)
	}

	data := []byte(cfg.Data)
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}

	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, errors.Wrap(err, "new http client")
	}

	qres, err := cli.ABCIQueryWithOptions(
		path, data, opts2)
	if err != nil {
		return nil, errors.Wrap(err, "querying")
	}

	return qres, nil
}

func handleBalanceQuery(path string, io commands.IO) string {
	// If the query is bank/balances & it contains a path, derive the address from the path
	if strings.Contains(path, "gno.land/") {
		i := strings.Index(path, balancesQuery)
		pkgPath := path[i+len(balancesQuery):]

		pkgAddr := gnolang.DerivePkgAddr(pkgPath)
		if len(pkgAddr.Bytes()) == 0 {
			io.Printf("could not derive address from path")
			return ""
		}

		path = balancesQuery + pkgAddr.String()
	}

	return path
}

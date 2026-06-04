package client

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

type QueryCfg struct {
	RootCfg *BaseCfg

	Data   string
	Path   string
	Height int64
	Prove  bool
}

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

	fs.Int64Var(
		&c.Height,
		"height",
		0,
		"query height",
	)

	fs.BoolVar(
		&c.Prove,
		"prove",
		false,
		"prove query result",
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

	data := []byte(cfg.Data)
	opts2 := client.ABCIQueryOptions{
		Height: cfg.Height,
		Prove:  cfg.Prove,
	}
	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, errors.Wrap(err, "new http client")
	}

	qres, err := cli.ABCIQueryWithOptions(
		context.Background(), cfg.Path, data, opts2)
	if err != nil {
		return nil, errors.Wrap(err, "querying")
	}

	return qres, nil
}

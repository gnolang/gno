package client

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/errors"
)

type queryCfg struct {
	rootCfg *baseCfg

	data   string
	height int64
	prove  bool

	// internal
	path string
}

func newQueryCmd(rootCfg *baseCfg) *commands.Command {
	cfg := &queryCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "query",
			ShortUsage: "query [flags] <path>",
			ShortHelp:  "Makes an ABCI query",
		},
		nil,
		func(_ context.Context, args []string) error {
			return execQuery(cfg, args)
		},
	)
}

func (c *queryCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.data,
		"data",
		"",
		"query data bytes",
	)

	fs.Int64Var(
		&c.height,
		"height",
		0,
		"query height (not yet supported)",
	)

	fs.BoolVar(
		&c.prove,
		"prove",
		false,
		"prove query result (not yet supported)",
	)
}

func execQuery(cfg *queryCfg, args []string) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	cfg.path = args[0]

	qres, err := queryHandler(cfg)
	if err != nil {
		return err
	}

	if qres.Response.Error != nil {
		fmt.Printf("Log: %s\n",
			qres.Response.Log)
		return qres.Response.Error
	}
	resdata := qres.Response.Data
	// XXX in general, how do we know what to show?
	// proof := qres.Response.Proof
	height := qres.Response.Height
	fmt.Printf("height: %d\ndata: %s\n",
		height,
		string(resdata))
	return nil
}

func queryHandler(cfg *queryCfg) (*ctypes.ResultABCIQuery, error) {
	remote := cfg.rootCfg.Remote
	if remote == "" || remote == "y" {
		return nil, errors.New("missing remote url")
	}

	data := []byte(cfg.data)
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(
		cfg.path, data, opts2)
	if err != nil {
		return nil, errors.Wrap(err, "querying")
	}

	return qres, nil
}

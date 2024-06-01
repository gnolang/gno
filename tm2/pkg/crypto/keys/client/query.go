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
	Output string
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

	fs.StringVar(
		&c.Output,
		"output",
		TEXT_FORMAT,
		"format of query's output",
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

	// If there is an error in the response, return the log message
	if qres.Response.Error != nil {
		io.Printf("Log: %s\n", qres.Response.Log)
		return qres.Response.Error
	}

	switch cfg.Output {
	case TEXT_FORMAT:
		io.Printf("height: %d\ndata: %s\n", qres.Response.Height, string(qres.Response.Data))
	case JSON_FORMAT:
		io.Printf(formatQueryResponse(qres.Response))
	default:
		return errors.New("Invalid output format")
	}

	return nil
}

func QueryHandler(cfg *QueryCfg) (*ctypes.ResultABCIQuery, error) {
	remote := cfg.RootCfg.Remote
	if remote == "" {
		return nil, errors.New("missing remote url")
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
		cfg.Path, data, opts2)
	if err != nil {
		return nil, errors.Wrap(err, "querying")
	}

	return qres, nil
}

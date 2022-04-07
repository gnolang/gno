package client

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

type QueryOptions struct {
	BaseOptions        // home,remote,...
	Data        []byte `flag:"data" help:"query data bytes"`                        // <pkgpath>\n<expr> for queryexprs.
	Height      int64  `flag:"height" help:"query height (not yet supported)"`      // not yet used
	Prove       bool   `flag:"prove" help:"prove query result (not yet supported)"` // not yet used

	// internal
	Path string `flag:"-"`
}

var DefaultQueryOptions = QueryOptions{
	BaseOptions: DefaultBaseOptions,
}

func queryApp(cmd *command.Command, args []string, iopts interface{}) error {
	var opts QueryOptions = iopts.(QueryOptions)

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: query <path>")
		return errors.New("invalid args")
	}
	opts.Path = args[0]

	qres, err := QueryHandler(opts)
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

func QueryHandler(opts QueryOptions) (*ctypes.ResultABCIQuery, error) {
	remote := opts.Remote
	if remote == "" || remote == "y" {
		return nil, errors.New("missing remote url")
	}

	data := opts.Data
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(
		opts.Path, data, opts2)
	if err != nil {
		return nil, errors.Wrap(err, "querying")
	}

	return qres, nil
}

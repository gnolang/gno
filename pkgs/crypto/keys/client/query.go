package client

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

type QueryOptions struct {
	BaseOptions        // home,remote,...
	Data        []byte `flag:"data" help:"query data bytes"`                        // <pkgpath>\n<expr> for queryexprs.
	Height      int64  `flag:"height" help:"query height (not yet supported)"`      // not yet used
	Prove       bool   `flag:"prove" help:"prove query result (not yet supported)"` // not yet used
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
	remote := opts.Remote
	if remote == "" || remote == "y" {
		return errors.New("missing remote url")
	}
	path := args[0]
	data := opts.Data
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	cli := client.NewHTTP(remote, "/websocket")
	qres, err := cli.ABCIQueryWithOptions(
		path, data, opts2)
	if err != nil {
		return errors.Wrap(err, "querying")
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

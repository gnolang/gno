package gnomod

import (
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

func queryChain(remote string, qpath string, data []byte) (res *abci.ResponseQuery, err error) {
	opts2 := client.ABCIQueryOptions{
		// Height: height, XXX
		// Prove: false, XXX
	}
	cli, err := client.NewHTTPClient(remote)
	if err != nil {
		return nil, err
	}

	qres, err := cli.ABCIQueryWithOptions(qpath, data, opts2)
	if err != nil {
		return nil, err
	}
	if qres.Response.Error != nil {
		fmt.Printf("Log: %s\n", qres.Response.Log)
		return nil, qres.Response.Error
	}

	return &qres.Response, nil
}

package client

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
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

	output := formatQueryResponse(qres.Response)
	io.Printf(output)

	return qres.Response.Error
}

func formatQueryResponse(res abci.ResponseQuery) string {
	if res.Error != nil {
		// If there is an error in the response, return the log message
		return fmt.Sprintf("Log: %s\n", res.Log)
	}

	// Default response string in case unmarshalling or marshalling fails
	defaultResponse := fmt.Sprintf("height: %d\ndata: %s\n", res.Height, string(res.Data))

	// Unmarshal the original response data into a json.RawMessage
	// This allows us to handle arbitrary JSON structures without knowing their schema
	var data json.RawMessage
	err := json.Unmarshal(res.Data, &data)
	if err != nil {
		return defaultResponse
	}

	// Create a struct to hold the final JSON structure with ordered fields
	formattedData := struct {
		Height int64           `json:"height"`
		Data   json.RawMessage `json:"data"`
	}{
		Height: res.Height,
		Data:   data,
	}

	// Marshal the final struct into an indented JSON string for readability
	formattedResponse, err := json.MarshalIndent(formattedData, "", "  ")
	if err != nil {
		return defaultResponse
	}

	// Return the formatted JSON string
	return string(formattedResponse)
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

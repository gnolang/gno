package client

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"

	"github.com/gnolang/gno/tm2/pkg/amino"
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

	if cfg.RootCfg.Json {
		return printResultABCIQueryJson(qres, io)
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

	defaultValues := url.Values{}
	if cfg.RootCfg.Json {
		defaultValues.Set("format", "json")
	}

	path, err := generatePathQuery(cfg.Path, defaultValues)
	if err != nil {
		return nil, errors.Wrap(err, "generate path query error")
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
		context.Background(), path, data, opts2)
	if err != nil {
		return nil, errors.Wrap(err, "querying")
	}

	return qres, nil
}

func printResultABCIQueryJson(qres *ctypes.ResultABCIQuery, io commands.IO) error {
	var output struct {
		Response json.RawMessage `json:"response"`
		Data     json.RawMessage `json:"data,omitempty"`
	}

	var err error
	if output.Response, err = amino.MarshalJSON(qres.Response); err != nil {
		io.ErrPrintfln("Unable to marshal response %+v\n", qres)
		return fmt.Errorf("amino marshal json error: %w", err)
	}

	data := qres.Response.Data
	switch {
	case len(data) == 0:
		output.Data = []byte(`[]`)
	case data[0] == '[', data[len(data)-1] == ']':
		fallthrough
	case data[0] == '{', data[len(data)-1] == '}':
		output.Data = data
	default:
		output.Data, _ = json.Marshal(qres.Response.Data)
	}

	var buff bytes.Buffer
	jqueryEnc := json.NewEncoder(&buff)
	jqueryEnc.SetEscapeHTML(false) // disable HTML escaping, as we want to correctly display `<`, `>`

	if err := jqueryEnc.Encode(output); err != nil {
		io.ErrPrintfln("Unable to marshal\n Response: %+v\n Data: %+v\n",
			string(output.Response),
			string(output.Data),
		)
		return fmt.Errorf("marshal json error: %w", err)
	}

	// Print out json.
	io.Printf(buff.String())

	if qres.Response.IsErr() {
		return commands.ExitCodeError(1)
	}

	return nil
}

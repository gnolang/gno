package keyscli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

func NewQueryJSONCmd(rootCfg *client.BaseCfg, io commands.IO) *commands.Command {
	cfg := &client.QueryCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "jquery",
			ShortUsage: "jquery [flags] <path>",
			ShortHelp:  "EXPERIMENTAL: makes an ABCI query and return a result in json",
			LongHelp:   "EXPERIMENTAL: makes an ABCI query and return a result in json",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execQuery(cfg, args, io)
		},
	)
}

func execQuery(cfg *client.QueryCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}

	cfg.Path = args[0]
	if cfg.Path == "vm/qeval" {
		// automatically add json suffix for qeval
		cfg.Path = cfg.Path + "/json"
	}

	qres, err := client.QueryHandler(cfg)
	if err != nil {
		return err
	}

	var output struct {
		Response json.RawMessage `json:"response"`
		Returns  json.RawMessage `json:"returns,omitempty"`
	}

	if output.Response, err = amino.MarshalJSONIndent(qres.Response, "", "  "); err != nil {
		io.ErrPrintfln("Unable to marshal response %+v\n", qres)
		return fmt.Errorf("amino marshal json error: %w", err)
	}

	// XXX: this is probably too specific
	if cfg.Path == "vm/qeval/json" {
		if len(qres.Response.Data) > 0 {
			output.Returns = qres.Response.Data
		} else {
			output.Returns = []byte("[]")
		}
	}

	res, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		io.ErrPrintfln("Unable to marshal output %+v\n", qres)
		return fmt.Errorf("marshal json error: %w", err)
	}

	io.Println(string(res))
	return nil
}

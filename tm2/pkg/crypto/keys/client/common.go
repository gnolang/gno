package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type BaseOptions struct {
	Home                  string
	Remote                string
	Quiet                 bool
	Json                  bool
	InsecurePasswordStdin bool
	Config                string
	// OnTxSuccess is called when the transaction tx succeeds. It can, for example,
	// print info in the result. If OnTxSuccess is nil, print basic info.
	OnTxSuccess func(tx std.Tx, res *ctypes.ResultBroadcastTxCommit)
}

var DefaultBaseOptions = BaseOptions{
	Home:                  "",
	Remote:                "127.0.0.1:26657",
	Quiet:                 false,
	Json:                  false,
	InsecurePasswordStdin: false,
	Config:                "",
}

func printJson(v any, io commands.IO) {
	res, err := json.Marshal(v)
	if err != nil {
		io.ErrPrintfln("unable to marshal value %q: %+v", err.Error())
		return
	}

	io.Println(string(res))
}

func generatePathQuery(path string, def url.Values) (string, error) {
	path, query, _ := strings.Cut(path, "?")
	values, err := url.ParseQuery(query)
	if err != nil {
		return "", fmt.Errorf("invalid path query %q: %w", query, err)
	}

	for k := range def {
		if values.Has(k) {
			continue
		}
		values.Set(k, def.Get(k))
	}

	if len(values) > 0 {
		path = path + "?" + values.Encode() // generate path
	}

	return path, nil
}

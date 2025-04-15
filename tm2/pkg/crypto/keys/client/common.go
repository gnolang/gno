package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type BaseOptions struct {
	Home                  string
	Remote                string
	Quiet                 bool
	Json                  bool
	InsecurePasswordStdin bool
	Config                string
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
	res, err := json.MarshalIndent(v, "", "\t")
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

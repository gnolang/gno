package client

import (
	"encoding/json"

	"github.com/gnolang/gno/tm2/pkg/amino"
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

func printAminoJson(v any, io commands.IO) {
	res, err := amino.MarshalJSONIndent(v, "", "\t")
	if err != nil {
		io.ErrPrintfln("unable to marshal value %q: %+v", err.Error())
		return
	}

	io.Println(string(res))
}

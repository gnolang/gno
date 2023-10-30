package main

import (
	"context"
	"flag"

	keysclient "github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type requiredOptions struct {
	keysclient.BaseOptions
	GasWanted       int64
	GasFee          string
	ChainID         string
	KeyNameOrBech32 string
	PkgPath         string
	Debug           bool
	Command         string
	CallAction      bool
	QueryAction     bool
}

func (opts *requiredOptions) flagSet() *flag.FlagSet {

	fs := flag.NewFlagSet("exec contract", flag.ExitOnError)
	defaultHome := keysclient.DefaultBaseOptions.Home

	fs.BoolVar(&opts.Debug, "debug", false, "verbose output")
	fs.Int64Var(&opts.GasWanted, "gas-wanted", 2000000, "gas requested for tx")
	fs.StringVar(&opts.GasFee, "gas-fee", "1000000ugnot", "gas payment fee")
	fs.StringVar(&opts.ChainID, "chainid", "staging", "")
	fs.StringVar(&opts.PkgPath, "pkgpath", defaultPkgPath, "blog realm path")
	fs.StringVar(&opts.KeyNameOrBech32, "key", "", "key name or bech32 address")

	fs.BoolVar(&opts.CallAction, "call", false, "call function")
	fs.BoolVar(&opts.QueryAction, "query", false, "query function")

	// keysclient.BaseOptions
	fs.StringVar(&opts.Home, "home", defaultHome, "home directory")
	fs.StringVar(&opts.Remote, "remote", defaultRemote, "remote node URL")
	fs.BoolVar(&opts.Quiet, "quiet", false, "for parsing output")
	fs.BoolVar(&opts.InsecurePasswordStdin, "insecure-password-stdin", false, "WARNING! take password from stdin")

	return fs
}

func Main2() {

	var reqOpts requiredOptions
	root := ffcli.Command{
		ShortUsage: "realm-name KEY COMMAND",
		ShortHelp:  "query or call a contract function",
		FlagSet:    reqOpts.flagSet(),
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 2 {
				return flag.ErrHelp
			}
			opts.KeyNameOrBech32 = args[0]
			posts := args[1:]
			return doPublish(ctx, posts, opts)
		},
	}

}

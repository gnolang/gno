package main

import (
	"context"
	"fmt"

	ff "github.com/peterbourgon/ff/v4"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

func (s *service) NewStartCmd() *ff.Command {
	rootFlags := ff.NewFlagSet("autocounterd")
	s.mnemonic = rootFlags.StringLong("mnemonic", "", "mnemonic")
	s.rpcURL = rootFlags.StringLong("rpc", "127.0.0.1:26657", "rpc url endpoint")

	cmd := &ff.Command{
		Name:  "start",
		Flags: rootFlags,
		Exec:  s.execStart,
	}

	return cmd
}

func (s *service) execStart(ctx context.Context, args []string) error {
	signer, err := gnoclient.SignerFromBip39(s.MustGetMnemonic(), "", uint32(0), uint32(0))
	if err != nil {
		return err
	}

	rpcClient := rpcclient.NewHTTP(s.MustGetRPC(), "/websocket")

	client := gnoclient.Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	res, err := client.Call(gnoclient.CallCfg{
		PkgPath:   "r/portalloop/counter",
		FuncName:  "Incr",
		GasFee:    "10000000ugnot",
		GasWanted: 800000,
	})
	if err != nil {
		return fmt.Errorf("failed to call Incr: %w", err)
	}
	_ = res
	fmt.Printf("res: %#v\n", res)

	return nil
}

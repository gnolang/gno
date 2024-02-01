package main

import (
	"context"
	"fmt"
	"time"

	ff "github.com/peterbourgon/ff/v4"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

func (s *service) NewStartCmd() *ff.Command {
	rootFlags := ff.NewFlagSet("autocounterd")
	s.rpcURL = rootFlags.StringLong("rpc", "127.0.0.1:26657", "rpc url endpoint")
	s.chainID = rootFlags.StringLong("chain-id", "dev", "chain-id")
	s.mnemonic = rootFlags.StringLong("mnemonic", "", "mnemonic")
	s.realmPath = rootFlags.StringLong("realm", "gno.land/r/portal/counter", "realm path")
	s.incrementInterval = rootFlags.DurationLong("interval", 15*time.Second, "Increment counter interval")

	cmd := &ff.Command{
		Name:  "start",
		Flags: rootFlags,
		Exec:  s.execStart,
	}

	return cmd
}

func (s *service) execStart(ctx context.Context, args []string) error {
	if err := s.Validate(); err != nil {
		return err
	}

	signer, err := gnoclient.SignerFromBip39(s.MustGetMnemonic(), s.MustGetChainID(), "", uint32(0), uint32(0))
	if err != nil {
		return err
	}
	if err := signer.Validate(); err != nil {
		return err
	}

	rpcClient := rpcclient.NewHTTP(s.MustGetRPC(), "/websocket")

	client := gnoclient.Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	for {
		res, err := client.Call(gnoclient.CallCfg{
			PkgPath:   s.MustGetRealmPath(),
			FuncName:  "Incr",
			GasFee:    "10000000ugnot",
			GasWanted: 800000,
			Args:      nil,
		})
		_ = res

		if err != nil {
			fmt.Printf("[ERROR] Failed to call Incr on %s, %+v\n", s.MustGetRealmPath(), err.Error())
		} else {
			fmt.Println("[INFO] Counter incremented with success")
		}
		time.Sleep(s.MustGetIncrementInterval())
	}
}

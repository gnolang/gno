package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type startCfg struct {
	rpcURL            string
	chainID           string
	mnemonic          string
	realmPath         string
	incrementInterval time.Duration
}

func (cfg *startCfg) Validate() error {
	switch {
	case cfg.rpcURL == "":
		return fmt.Errorf("rpc url cannot be empty")
	case cfg.chainID == "":
		return fmt.Errorf("chainID cannot be empty")
	case cfg.mnemonic == "":
		return fmt.Errorf("mnemonic cannot be empty")
	case cfg.realmPath == "":
		return fmt.Errorf("realmPath cannot be empty")
	case cfg.incrementInterval == 0:
		return fmt.Errorf("interval cannot be empty")
	}

	return nil
}

func NewStartCmd(io commands.IO) *commands.Command {
	cfg := &startCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "start",
			ShortUsage: "start [flags]",
			ShortHelp:  "Runs the linter for the specified packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execStart(cfg, args, io)
		},
	)
}

func (c *startCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.rpcURL, "rpc", "127.0.0.1:26657", "rpc url endpoint")
	fs.StringVar(&c.chainID, "chain-id", "dev", "chain-id")
	fs.StringVar(&c.mnemonic, "mnemonic", "", "mnemonic")
	fs.StringVar(&c.realmPath, "realm", "gno.land/r/portal/counter", "realm path")
	fs.DurationVar(&c.incrementInterval, "interval", 15*time.Second, "Increment counter interval")
}

func execStart(cfg *startCfg, args []string, io commands.IO) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	signer, err := gnoclient.SignerFromBip39(cfg.mnemonic, cfg.chainID, "", uint32(0), uint32(0))
	if err != nil {
		return err
	}
	if err := signer.Validate(); err != nil {
		return err
	}

	rpcClient := rpcclient.NewHTTP(cfg.rpcURL, "/websocket")

	client := gnoclient.Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	for {
		res, err := client.Call(gnoclient.CallCfg{
			PkgPath:   cfg.realmPath,
			FuncName:  "Incr",
			GasFee:    "10000000ugnot",
			GasWanted: 800000,
			Args:      nil,
		})
		_ = res

		if err != nil {
			fmt.Printf("[ERROR] Failed to call Incr on %s, %+v\n", cfg.realmPath, err.Error())
		} else {
			fmt.Println("[INFO] Counter incremented with success")
		}
		time.Sleep(cfg.incrementInterval)
	}
}

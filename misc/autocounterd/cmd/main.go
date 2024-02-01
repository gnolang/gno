package main

import (
	"context"
	"fmt"
	"os"
	"time"

	ff "github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

type service struct {
	rpcURL            *string
	chainID           *string
	mnemonic          *string
	realmPath         *string
	incrementInterval *time.Duration
}

func (s service) Validate() error {
	if s.mnemonic != nil && *s.mnemonic == "" {
		return fmt.Errorf("mnemonic is missing")
	} else if s.rpcURL != nil && *s.rpcURL == "" {
		return fmt.Errorf("rpc url is missing")
	} else if s.chainID != nil && *s.chainID == "" {
		return fmt.Errorf("chain_id is missing")
	} else if s.incrementInterval == nil {
		return fmt.Errorf("interval is missing")
	} else if s.realmPath != nil && *s.realmPath == "" {
		return fmt.Errorf("realm path is missing")
	}
	return nil
}

func (s service) MustGetRPC() string                      { return *s.rpcURL }
func (s service) MustGetChainID() string                  { return *s.chainID }
func (s service) MustGetMnemonic() string                 { return *s.mnemonic }
func (s service) MustGetRealmPath() string                { return *s.realmPath }
func (s service) MustGetIncrementInterval() time.Duration { return *s.incrementInterval }

func main() {
	s := &service{}

	rootCmd := &ff.Command{
		Name: "autocounterd",
		Subcommands: []*ff.Command{
			s.NewStartCmd(),
		},
	}

	err := rootCmd.ParseAndRun(
		context.Background(),
		os.Args[1:],
		ff.WithEnvVarPrefix("COUNTER"))

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Command(rootCmd))
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

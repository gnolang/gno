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
	mnemonic          *string
	rpcURL            *string
	incrementInterval *time.Duration
}

func (s service) MustGetMnemonic() string {
	if s.mnemonic != nil && *s.mnemonic != "" {
		return *s.mnemonic
	}
	panic("mnemonic is empty")
}

func (s service) MustGetRPC() string {
	if s.rpcURL != nil && *s.rpcURL != "" {
		return *s.rpcURL
	}
	panic("rpc url is empty")
}

func (s service) MustGetIncrementInterval() time.Duration {
	if s.incrementInterval != nil {
		return *s.incrementInterval
	}
	panic("increment interval is empty")
}

func main() {
	s := &service{}

	rootCmd := &ff.Command{
		Name: "autocounterd",
		// Flags: rootFlags,
		Subcommands: []*ff.Command{
			s.NewStartCmd(),
			s.NewDeployCmd(),
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

package main

import (
	"context"
	"fmt"
	"os"

	ff "github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

type service struct {
	mnemonic *string
	rpcURL   *string
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

func main() {
	s := &service{}
	// rootFlags := ff.NewFlagSet("autocounterd")
	// s.mnemonic = rootFlags.StringLong("mnemonic", "", "mnemonic")

	rootCmd := &ff.Command{
		Name: "autocounterd",
		// Flags: rootFlags,
		Subcommands: []*ff.Command{
			s.NewStartCmd(),
			s.NewDeployCmd(),
		},
	}

	if err := rootCmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Command(rootCmd))
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

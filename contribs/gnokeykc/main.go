package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/zalando/go-keyring"
)

func main() {
	stdio := commands.NewDefaultIO()
	wrappedio := &wrappedIO{IO: stdio}
	cmd := keyscli.NewRootCmd(wrappedio)
	cmd.AddSubCommands(newKcCmd(stdio))

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)

		os.Exit(1)
	}
}

type wrappedIO struct {
	commands.IO
}

func (io *wrappedIO) GetPassword(prompt string, insecure bool) (string, error) {
	return keyring.Get(kcService, kcName)
}

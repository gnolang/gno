package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/zalando/go-keyring"
)

func main() {
	stdio := commands.NewDefaultIO()
	wrappedio := &wrappedIO{IO: stdio}
	cmd := client.NewRootCmd(wrappedio)
	cmd.AddSubCommands(newKcCmd(stdio))

	cmd.Execute(context.Background(), os.Args[1:])
}

type wrappedIO struct {
	commands.IO
}

func (io *wrappedIO) GetPassword(prompt string, insecure bool) (string, error) {
	return keyring.Get(kcService, kcName)
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/crypto/keys/client"
)

func main() {
	cmd := client.NewRootCmd()

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

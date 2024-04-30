package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := NewStartCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}

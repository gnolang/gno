package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/integration"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := integration.RunMain(ctx, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

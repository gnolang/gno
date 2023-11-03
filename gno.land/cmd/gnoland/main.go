package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

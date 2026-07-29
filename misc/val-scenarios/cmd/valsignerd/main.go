package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gnolang/gno/misc/val-scenarios/pkg/valsigner"
)

func main() {
	var (
		keyFile    = flag.String("key-file", "", "path to the validator private key JSON file")
		control    = flag.String("listen-addr", ":8080", "HTTP control API listen address")
		remoteAddr = flag.String("remote-signer-addr", "tcp://0.0.0.0:26659", "remote signer listen address")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	server, err := valsigner.NewServer(*keyFile, *control, *remoteAddr, logger)
	if err != nil {
		logger.Error("unable to create signer server", "err", err)
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		logger.Error("unable to start signer server", "err", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	if err := server.Stop(); err != nil {
		logger.Error("unable to stop signer server", "err", err)
		os.Exit(1)
	}
}

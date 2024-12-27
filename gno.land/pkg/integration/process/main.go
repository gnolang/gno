package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/integration"
)

const gracefulShutdown = time.Second * 5

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Read the configuration from standard input
	configData, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error reading stdin: %v", err)
	}

	// Unmarshal the JSON configuration
	var cfg integration.ProcessNodeConfig
	if err := json.Unmarshal(configData, &cfg); err != nil {
		log.Fatalf("error unmarshaling JSON: %v", err)
	}

	// Run the node
	ccErr := make(chan error, 1)
	go func() {
		ccErr <- integration.RunNode(ctx, &cfg)
		close(ccErr)
	}()

	// Wait for the node gracefully terminate
	<-ctx.Done()

	// Attempt graceful shutdown
	select {
	case <-time.After(gracefulShutdown):
		log.Fatalf("unable to gracefully stop the node, exiting now")
	case err := <-ccErr: // done
		if err != nil {
			log.Fatalf("unable to gracefully stop the node: %v", err)
		}
	}
}

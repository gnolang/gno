package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gnostats/proto"
)

// config holds both the injected Gno node RPC client and the Hub gRPC client.
type config struct {
	hClient      proto.HubClient
	rClient      rpcClient
	pollInterval time.Duration // Minimum time interval between two data points
}

type agent struct {
	cfg    config
	cancel context.CancelFunc
}

// Start registers with the Hub using Gno node static info, then pushes dynamic
// info from the Gno node to the Hub at intervals specified by pollInterval
func (a *agent) Start(ctx context.Context) error {
	// Store a cancelFunc to make the agent stoppable using the Stop() method
	ctx, a.cancel = context.WithCancel(ctx)
	defer a.cancel()

	collector := NewCollector(a.cfg.rClient)

	// Get static info from the Gno node
	staticInfo, err := collector.CollectStatic(ctx)
	if err != nil {
		return fmt.Errorf("unable to collect static info: %w", err)
	}

	// Register with the Hub using static info
	if _, err = a.cfg.hClient.Register(ctx, staticInfo); err != nil {
		return fmt.Errorf("unable to register with hub: %w", err)
	}

	// Get the Hub data point stream
	stream, err := a.cfg.hClient.PushData(ctx)
	if err != nil {
		return fmt.Errorf("unable to get data stream: %w", err)
	}

	// Push data points until the context is done
	for {
		// Get dynamic info from the Gno node
		dynamicInfo, err := collector.CollectDynamic(ctx)
		if err != nil {
			return fmt.Errorf("unable to collect dynamic info: %w", err)
		}

		// Push dynamic info to the Hub stream
		if err = stream.Send(dynamicInfo); err != nil {
			return fmt.Errorf("unable to send dynamic info: %w", err)
		}

		select {
		case <-time.After(a.cfg.pollInterval):
			// Wait for the specified interval before pushing a new data point
		case <-ctx.Done():
			return nil
		}
	}
}

// Stop stops the agent
func (a *agent) Stop() {
	a.cancel()
}

// NewAgent creates a new agent using the provided config
func NewAgent(cfg config) *agent {
	return &agent{
		cfg: cfg,
	}
}

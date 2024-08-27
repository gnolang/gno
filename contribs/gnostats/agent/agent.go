package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gnostats/proto"
)

// angent holds both the injected Gno node RPC client and the Hub gRPC client.
type agent struct {
	hClient      proto.HubClient
	rClient      rpcClient
	pollInterval time.Duration // Minimum time interval between two data points
}

type Option func(*agent)

// WithPollInterval sets the agent poll interval between two data points
func WithPollInterval(interval time.Duration) Option {
	return func(c *agent) {
		c.pollInterval = interval
	}
}

// Start registers with the Hub using Gno node static info, then pushes dynamic
// info from the Gno node to the Hub at intervals specified by pollInterval
func (a *agent) Start(ctx context.Context) error {
	collector := NewCollector(a.rClient)

	// Get static info from the Gno node
	staticInfo, err := collector.CollectStatic(ctx)
	if err != nil {
		return fmt.Errorf("unable to collect static info: %w", err)
	}

	// Register with the Hub using static info
	if _, err = a.hClient.Register(ctx, staticInfo); err != nil {
		return fmt.Errorf("unable to register with hub: %w", err)
	}

	// Get the Hub data point stream
	stream, err := a.hClient.PushData(ctx)
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
		case <-time.After(a.pollInterval):
			// Wait for the specified interval before pushing a new data point
		case <-ctx.Done():
			return nil
		}
	}
}

// NewAgent creates a new agent using the provided clients and options
func NewAgent(hClient proto.HubClient, rClient rpcClient, options ...Option) *agent {
	const defaultInverval = time.Second

	a := &agent{
		hClient:      hClient,
		rClient:      rClient,
		pollInterval: defaultInverval,
	}

	for _, opt := range options {
		opt(a)
	}

	return a
}

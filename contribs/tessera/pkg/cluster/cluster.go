package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gnolang/gno/contribs/tessera/pkg/cluster/docker"
)

type Cluster struct {
	logger *slog.Logger

	config  Config
	manager *docker.ClusterManager
}

func New(
	ctx context.Context,
	logger *slog.Logger,
	config Config,
	gnoRoot string,
) (*Cluster, error) {
	c := &Cluster{
		config: config,
		logger: logger,
	}

	// Create the Docker container manager.
	// This process directly interfaces with Docker on the machine,
	// and handles container processes for the Cluster
	clusterName := fmt.Sprintf(
		"cluster-%d",
		time.Now().Nanosecond(),
	)
	manager, err := docker.NewClusterManager(
		clusterName,
		logger.With("process", "cluster"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create Docker manager: %w", err)
	}

	// Build the docker images for the cluster
	if err = manager.BuildDockerfile(ctx, gnoRoot); err != nil {
		return nil, fmt.Errorf("unable to build Dockerfile: %w", err)
	}

	// Create a shared network
	if err = manager.SetupNetwork(ctx); err != nil {
		return nil, fmt.Errorf("unable to setup shared network: %w", err)
	}

	// Create a shared volume (primarily for genesis access)
	if err = manager.SetupSharedVolume(ctx); err != nil {
		return nil, fmt.Errorf("unable to setup shared volume: %w", err)
	}

	// Create the cluster containers
	// TODO
	logger.Debug("Creating clusters....")

	// Create the genesis.json (shared)
	// TODO
	logger.Debug("Creating genesis.json....")

	c.manager = manager

	return c, nil
}

// Start starts the node cluster
func (c *Cluster) Start(ctx context.Context) error {
	// TODO

	return nil
}

func (c *Cluster) Shutdown(ctx context.Context) {
	// Stop the Docker containers, if any
	if c.manager != nil {
		c.manager.Cleanup(ctx)
	}
}

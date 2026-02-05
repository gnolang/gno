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
	imageName, err := manager.BuildDockerfile(ctx, gnoRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to build Dockerfile: %w", err)
	}

	// Create a shared network
	if err = manager.SetupNetwork(ctx); err != nil {
		return nil, fmt.Errorf("unable to setup shared network: %w", err)
	}

	// Create a shared volume (primarily for genesis access)
	sharedVolume, err := manager.SetupSharedVolume(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to setup shared volume: %w", err)
	}

	// Create the cluster containers
	// TODO
	logger.Debug("Creating clusters....")

	// Create the genesis.json (shared)
	logger.Debug("Creating genesis.json....")

	// To generate the genesis.json, a container for this purpose
	// will be created and used
	genesisContainerCfg := docker.ContainerConfig{
		Name:       fmt.Sprintf("%s-genesis", clusterName),
		Image:      imageName,
		Entrypoint: []string{"/bin/sh", "-c", "sleep infinity"},
		Cmd:        nil, // TODO check
	}

	if err = manager.CreateContainer(ctx, genesisContainerCfg); err != nil {
		return nil, fmt.Errorf("unable to create genesis container: %w", err)
	}

	// Create an empty genesis.json
	genesisPath := fmt.Sprintf("%s/genesis.json", sharedVolume)

	execOutput, err := manager.ExecuteCmd(
		ctx,
		genesisContainerCfg.Name,
		[]string{
			fmt.Sprintf(
				"gnogenesis generate -chain-id %s -genesis-time %d -output-path %s",
				config.Genesis.ChainID,
				config.Genesis.GenesisTime.Unix(),
				genesisPath,
			),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to execute command on container: %w", err)
	}

	// TODO cleanup
	fmt.Println(execOutput)

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

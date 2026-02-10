package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerHandler struct {
	DockerClient *client.Client
	Logger       *slog.Logger
}

const (
	gnoContainerLabel string = "gno-staging"
	gnoOfficialImage  string = "ghcr.io/gnolang/gno/gnoland:master"
)

// Checks if a fresh pull of Master image corresponds to a new version
func (dh *DockerHandler) CheckPulledMasterImage(ctx context.Context) (bool, error) {
	localImage, _, err := dh.DockerClient.ImageInspectWithRaw(ctx, gnoOfficialImage)
	if err != nil {
		dh.DockerClient.ImagePull(ctx, gnoOfficialImage, types.ImagePullOptions{})
		return true, nil
	}

	// Get local image digest
	if len(localImage.RepoDigests) == 0 {
		// Assume it's locally built and not pulled
		return true, nil
	}
	// local digest include full repository name
	localDigestPrefix := strings.ReplaceAll(gnoOfficialImage, ":master", "")
	localDigest := strings.ReplaceAll(localImage.RepoDigests[0], fmt.Sprintf("%s@", localDigestPrefix), "")

	// Get remote image digest
	remoteImage, err := dh.DockerClient.DistributionInspect(ctx, gnoOfficialImage, "")
	if err != nil {
		return false, err
	}
	remoteDigest := remoteImage.Descriptor.Digest.String()
	return localDigest != remoteDigest, nil
}

// Gather the list of current running containter of
func (dh *DockerHandler) GetActiveGnoPortalLoopContainers(ctx context.Context) ([]types.Container, error) {
	containers, err := dh.DockerClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return []types.Container{}, err
	}
	gnoPortalLoopContainers := make([]types.Container, 0)

	for _, container := range containers {
		if _, exists := container.Labels[gnoContainerLabel]; exists {
			gnoPortalLoopContainers = append(gnoPortalLoopContainers, container)
		}
	}

	return gnoPortalLoopContainers, nil
}

// Starts the Gno Portal Loop Container
func (dh *DockerHandler) StartGnoPortalLoopContainer(ctx context.Context, containerName, hostPwd string, pullImage bool) (*types.Container, error) {
	// Create Docker volume
	volume, err := dh.DockerClient.VolumeCreate(ctx, volume.CreateOptions{
		Name: containerName,
	})
	if err != nil {
		return nil, err
	}

	// force pull image
	if pullImage {
		pullOutput, err := dh.DockerClient.ImagePull(ctx, gnoOfficialImage, types.ImagePullOptions{})
		if err != nil {
			return nil, err
		}
		defer pullOutput.Close()

		// Read until EOF to ensure pull completes
		_, err = io.Copy(io.Discard, pullOutput)
		if err != nil {
			return nil, fmt.Errorf("failed to read image pull output: %w", err)
		}
	}

	// Run Docker container
	dockerContainer, err := dh.DockerClient.ContainerCreate(ctx, &container.Config{
		Image: gnoOfficialImage,
		Labels: map[string]string{
			gnoContainerLabel: containerName,
		},
		WorkingDir: "/gnoroot",
		Env: []string{
			"MONIKER=the-staging-chain",
			"GENESIS_BACKUP_FILE=/backups/backup.jsonl",
			"GENESIS_BALANCES_FILE=/backups/balances.jsonl",
			"FAUCET_ADDRESS=g1qhuef2450xh7g7na8s865nreu2xw8j84kgkvt5",
		},
		Entrypoint: []string{"/scripts/start.sh"},
		ExposedPorts: nat.PortSet{
			"26656/tcp": struct{}{},
			"26657/tcp": struct{}{},
		},
	}, &container.HostConfig{
		// This is probably no more supported by Docker and was left just as historical reference
		// but should be useless for the application to succeed
		// PublishAllPorts: true,
		PortBindings: nat.PortMap{
			"26657/tcp": []nat.PortBinding{
				{HostIP: "127.0.0.1"},
			},
		},
		Binds: []string{
			fmt.Sprintf("%s/scripts:/scripts", hostPwd),
			fmt.Sprintf("%s/backups:/backups", hostPwd),
			fmt.Sprintf("%s:/gnoroot/gnoland-data", volume.Name),
		},
	}, nil, nil, containerName)
	if err != nil {
		return nil, err
	}

	err = dh.DockerClient.NetworkConnect(ctx, "portal-loop", dockerContainer.ID, nil)
	if err != nil {
		return nil, err
	}

	if err := dh.DockerClient.ContainerStart(ctx, dockerContainer.ID, container.StartOptions{}); err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 5)

	containers, err := dh.GetActiveGnoPortalLoopContainers(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if c.ID == dockerContainer.ID {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("container not found")
}

// Gather the RPC Port published for the given container
func (dh *DockerHandler) GetPublishedRPCPort(dockerContainer types.Container) string {
	for _, p := range dockerContainer.Ports {
		if p.Type == "tcp" && p.PrivatePort == uint16(26657) {
			ip := dockerContainer.NetworkSettings.Networks["portal-loop"].IPAddress
			return fmt.Sprintf("http://%s:%d", ip, int(p.PrivatePort))
		}
	}
	return ""
}

// Removes the given containers and the linked volumes
func (dh *DockerHandler) RemoveContainersWithVolumes(ctx context.Context, containers []types.Container) error {
	for _, c := range containers {
		dh.Logger.Info("removing container",
			slog.String("container.id", c.ID),
			slog.Any("container.ports", c.Ports),
		)
		err := dh.DockerClient.ContainerRemove(ctx, c.ID, container.RemoveOptions{
			Force:         true,  // Force the removal of a running container
			RemoveVolumes: true,  // Remove the volumes associated with the container
			RemoveLinks:   false, // Remove the specified link and not the underlying container
		})
		if err != nil {
			return err
		}
		for _, mount := range c.Mounts {
			if mount.Type == "volume" {
				err = dh.DockerClient.VolumeRemove(ctx, mount.Name, true)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

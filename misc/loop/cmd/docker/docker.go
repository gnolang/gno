package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

type DockerHandler struct {
	DockerClient *client.Client
}

const GnoOfficialImage string = "ghcr.io/gnolang/gno/gnoland:master"

// Checks if a fresh pull of Master image corresponds to a new version
func (dh *DockerHandler) CheckPulledMasterImage(ctx context.Context) (bool, error) {
	reader, err := dh.DockerClient.ImagePull(ctx, GnoOfficialImage, types.ImagePullOptions{})
	if err != nil {
		return false, err
	}
	defer reader.Close()

	var b bytes.Buffer
	_, err = io.Copy(&b, reader)
	if err != nil {
		return false, err
	}

	return !strings.Contains(b.String(), "Image is up to date"), nil
}

// Gather the list of current running containter of
func (dh *DockerHandler) GetActiveGnoPortalLoopContainers(ctx context.Context) ([]types.Container, error) {
	containers, err := dh.DockerClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return []types.Container{}, err
	}
	gnoPortalLoopContainers := make([]types.Container, 0)

	for _, container := range containers {
		if _, exists := container.Labels["the-portal-loop"]; exists {
			gnoPortalLoopContainers = append(gnoPortalLoopContainers, container)
		}
	}

	return gnoPortalLoopContainers, nil
}

// Starts the Gno Portal Loop Container
func (dh *DockerHandler) StartGnoPortalLoopContainer(ctx context.Context, containerName string, hostPwd string) (*types.Container, error) {
	// Create Docker volume
	volume, err := dh.DockerClient.VolumeCreate(ctx, volume.CreateOptions{
		Name: containerName,
	})
	if err != nil {
		return nil, err
	}

	// Run Docker container
	dockerContainer, err := dh.DockerClient.ContainerCreate(ctx, &container.Config{
		Image: "ghcr.io/gnolang/gno/gnoland:master",
		Labels: map[string]string{
			"the-portal-loop": containerName,
		},
		WorkingDir: "/gnoroot",
		Env: []string{
			"MONIKER=the-portal-loop",
			"GENESIS_BACKUP_FILE=/backups/backup.jsonl",
			"GENESIS_BALANCES_FILE=/backups/balances.jsonl",
		},
		Entrypoint: []string{"/scripts/start.sh"},
		ExposedPorts: nat.PortSet{
			"26656/tcp": struct{}{},
			"26657/tcp": struct{}{},
		},
	}, &container.HostConfig{
		PublishAllPorts: true,
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

	if err := dh.DockerClient.ContainerStart(ctx, dockerContainer.ID, types.ContainerStartOptions{}); err != nil {
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

// Gether the RPC Port published for the given container
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
		logrus.WithFields(logrus.Fields{
			"container.id":    c.ID,
			"container.ports": c.Ports,
		}).Infof("remove container")
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

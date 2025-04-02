package docker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// TODO embed the dockerfile
var dockerFile []byte

var images = []string{
	"gnoland",
	"gnocontribs",
}

const (
	networkNamePrefix = "gnoland-network"
	volumeNamePrefix  = "gnoland-shared-volume"
)

type ClusterManager struct {
	client *client.Client
	logger *slog.Logger

	containerInfos map[string]ContainerInfo

	name string // unique cluster name
}

type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	HostIP  string
	Ports   map[string]string // container port -> host port
	Running bool
}

type ContainerConfig struct {
	Name       string            // unique container name
	Image      string            // base image for the container (built / provided)
	Ports      map[string]string // container port -> host port
	Entrypoint []string
	Cmd        []string
}

func NewClusterManager(logger *slog.Logger) (*ClusterManager, error) {
	// Create the Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create Docker client: %w", err)
	}

	return &ClusterManager{
		client: cli,
		logger: logger,
	}, nil
}

func (cm *ClusterManager) BuildImages(ctx context.Context) error {
	for _, img := range images {
		cm.logger.Debug("Building image", "image", img)
	}

	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerFile)),
	}

	err := tw.WriteHeader(tarHeader)
	if err != nil {
		return err
	}

	_, err = tw.Write(dockerFile)
	if err != nil {
		return err
	}

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	imageBuildRes, err := cm.client.ImageBuild(
		ctx,
		dockerFileTarReader,
		types.ImageBuildOptions{
			Context:    dockerFileTarReader,
			Dockerfile: "Dockerfile",
			Remove:     true,
		},
	)
	if err != nil {
		return err
	}

	// TODO copy response body?

	return imageBuildRes.Body.Close() // TODO critical error?
}

func (cm *ClusterManager) SetupNetwork(ctx context.Context) (string, error) {
	networkName := fmt.Sprintf("%s-%s", networkNamePrefix, cm.name)

	networks, err := cm.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, nw := range networks {
		if nw.Name == networkName {
			return nw.ID, nil
		}
	}

	resp, err := cm.client.NetworkCreate(
		ctx,
		networkName,
		network.CreateOptions{
			Driver: "bridge",
		},
	)
	if err != nil {
		return "", err
	}

	// TODO log ID

	return resp.ID, nil
}

func (cm *ClusterManager) SetupSharedVolume(ctx context.Context) (string, error) {
	volumes, err := cm.client.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return "", err
	}

	volumeName := fmt.Sprintf("%s-%s", volumeNamePrefix, cm.name)

	for _, vol := range volumes.Volumes {
		if vol.Name == volumeName {
			return vol.Mountpoint, nil
		}
	}

	vol, err := cm.client.VolumeCreate(ctx, volume.CreateOptions{
		Driver: "local",
		Name:   volumeName,
	})
	if err != nil {
		return "", err
	}

	return vol.Mountpoint, nil
}

func (cm *ClusterManager) CreateContainer(ctx context.Context) (string, error) {
	fmt.Printf("Creating container %s from image %s...\n", config.Name, config.Image)

	// Create port bindings
	var (
		portBindings = nat.PortMap{}
		exposedPorts = nat.PortSet{}
	)

	for containerPort, hostPort := range config.Ports {
		port := nat.Port(containerPort)
		exposedPorts[port] = struct{}{}

		// If host port is empty, it will be auto-assigned
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	// Set up volume mounts
	var mounts []mount.Mount

	// Add shared volume mount
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeVolume,
		Source: volumeName,
		Target: volumeMountPath,
	})

	// Create container
	resp, err := cm.client.ContainerCreate(
		ctx,
		&container.Config{
			Image:        config.Image,
			ExposedPorts: exposedPorts,
			Entrypoint:   config.Entrypoint,
			Cmd:          config.Cmd,
			Tty:          true,
		},
		&container.HostConfig{
			PortBindings: portBindings,
			Mounts:       mounts,
			AutoRemove:   false,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkName: {
					NetworkID: cm.networkID,
				},
			},
		},
		nil,
		config.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	containerID := resp.ID

	// Start container
	if err := cm.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	// Get container details
	containerJSON, err := cm.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return containerID, fmt.Errorf("container started but failed to inspect: %v", err)
	}

	// Get assigned ports
	assignedPorts := make(map[string]string)
	for containerPort, bindings := range containerJSON.NetworkSettings.Ports {
		if len(bindings) > 0 {
			assignedPorts[string(containerPort)] = bindings[0].HostPort
		}
	}

	// Save container info
	cm.containerInfos[config.Name] = ContainerInfo{
		ID:      containerID,
		Name:    config.Name,
		Image:   config.Image,
		HostIP:  "127.0.0.1", // TODO change?
		Ports:   assignedPorts,
		Running: true,
	}

	fmt.Printf("Container %s created with ID %s\n", config.Name, containerID)
	fmt.Printf("Ports: %v\n", assignedPorts)

	return containerID, nil
}

func (cm *ClusterManager) ExecuteCmd(
	ctx context.Context,
	containerName string,
	cmd []string,
) (string, error) {
	containerInfo, exists := cm.containerInfos[containerName]

	if !exists {
		return "", fmt.Errorf("container %s not found", containerName)
	}

	// Create exec configuration
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create exec instance
	execID, err := cm.client.ContainerExecCreate(ctx, containerInfo.ID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec instance: %v", err)
	}

	// Start exec instance
	resp, err := cm.client.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec instance: %v", err)
	}
	defer resp.Close()

	// Read the output
	var outBuf strings.Builder

	scanner := bufio.NewScanner(resp.Reader)
	for scanner.Scan() {
		line := scanner.Text()

		outBuf.WriteString(line)
		outBuf.WriteString("\n")
	}

	// Check for exec completion
	execInspect, err := cm.client.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return outBuf.String(), fmt.Errorf("failed to inspect exec instance: %v", err)
	}

	if execInspect.ExitCode != 0 {
		return outBuf.String(), fmt.Errorf("command exited with code %d", execInspect.ExitCode)
	}

	return outBuf.String(), nil
}

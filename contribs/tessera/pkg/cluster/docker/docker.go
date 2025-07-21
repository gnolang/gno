package docker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	networkNamePrefix = "gnoland-network"
	volumeNamePrefix  = "gnoland-shared-volume"
)

type ClusterManager struct {
	client *client.Client // docker API client
	logger *slog.Logger

	containerInfos map[string]ContainerInfo //

	name string // unique cluster name

	sharedVolumePath string // the shared volume mount point
	networkID        string // the common Docker network ID
}

type ContainerInfo struct {
	ID     string            // Docker-generated ID
	Name   string            // unique container name
	Image  string            // base image for the container
	HostIP string            // the host IP (localhost)
	Ports  map[string]string // container port -> host port
}

type ContainerConfig struct {
	Name       string            // unique container name
	Image      string            // base image for the container (built / provided)
	Ports      map[string]string // container port -> host port
	Entrypoint []string
	Cmd        []string
}

func NewClusterManager(name string, logger *slog.Logger) (*ClusterManager, error) {
	// Create the Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create Docker client: %w", err)
	}

	return &ClusterManager{
		client:         cli,
		logger:         logger,
		name:           name,
		containerInfos: make(map[string]ContainerInfo),
	}, nil
}

// BuildDockerfile builds the tessera/gnoland Dockerfile
// TODO separate this out from the cluster manager. Image building should be a separate process from
// cluster creation and management
func (cm *ClusterManager) BuildDockerfile(ctx context.Context, gnoRoot string) (string, error) {
	var (
		buf = bytes.NewBuffer(nil)
		tw  = tar.NewWriter(buf)
	)

	// Add the gno root folder to the Docker build context
	if err := addFilesToTar(tw, gnoRoot); err != nil {
		return "", err
	}

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerFile)),
	}

	err := tw.WriteHeader(tarHeader)
	if err != nil {
		return "", err
	}

	_, err = tw.Write(dockerFile)
	if err != nil {
		return "", err
	}

	// Add the .dockerignore
	ignoreHeader := &tar.Header{
		Name: ".dockerignore",
		Size: int64(len(dockerIgnore)),
		Mode: 0o644,
	}

	if err = tw.WriteHeader(ignoreHeader); err != nil {
		return "", err
	}

	if _, err = tw.Write(dockerIgnore); err != nil {
		return "", err
	}

	if err = tw.Close(); err != nil {
		return "", err
	}

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	imageName := "tessera/gnoland:latest"
	imageBuildRes, err := cm.client.ImageBuild(
		ctx,
		dockerFileTarReader,
		types.ImageBuildOptions{
			Context:     dockerFileTarReader,
			Dockerfile:  "Dockerfile",
			Remove:      true,
			ForceRemove: true,
			Tags:        []string{imageName},
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to build images from Dockerfile: %w", err)
	}

	// TODO remove
	scanner := bufio.NewScanner(imageBuildRes.Body)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		cm.logger.Error("unable to gracefully close scanner", "err", err)
	}

	if err = imageBuildRes.Body.Close(); err != nil {
		cm.logger.Warn(
			"unable to gracefully close image builder",
			"err", err,
		)
	}

	// Clean up dangling images
	_, err = cm.client.ImagesPrune(ctx, filters.NewArgs(filters.Arg("dangling", "true")))
	if err != nil {
		cm.logger.Warn("failed to prune dangling images", "err", err)
	}

	return imageName, nil
}

// addFilesToTar adds all files from the source directory to the tar writer
func addFilesToTar(tw *tar.Writer, srcDir string) error {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a relative path within the tar archive
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip if this is the source directory itself
		if relPath == "." {
			return nil
		}

		// Create the full path within the tar
		tarPath := relPath

		// For directories create an entry
		if info.IsDir() {
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}

			header.Name = tarPath + "/"
			header.Mode = int64(info.Mode())

			return tw.WriteHeader(header)
		}

		// For files, add the content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = tarPath
		header.Mode = int64(info.Mode())
		header.Size = info.Size()

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		_, err = io.Copy(tw, file)

		return err
	}

	return filepath.Walk(srcDir, walkFn)
}

// SetupNetwork creates the cluster's Docker network, if it doesn't exist
func (cm *ClusterManager) SetupNetwork(ctx context.Context) error {
	networkName := cm.networkName()

	// Check if the network exists already
	networks, err := cm.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to query Docker networks: %w", err)
	}

	for _, nw := range networks {
		if nw.Name == networkName {
			// Save the ID
			cm.networkID = nw.ID

			return nil
		}
	}

	// Network doesn't exist, so it needs to be created
	resp, err := cm.client.NetworkCreate(
		ctx,
		networkName,
		network.CreateOptions{
			Driver: "bridge",
		},
	)
	if err != nil {
		return fmt.Errorf("unable to create Docker network: %w", err)
	}

	// Save the ID
	cm.networkID = resp.ID

	return nil
}

// SetupSharedVolume creates a shared Docker volume for the cluster,
// for common artifact exchange (like genesis files)
func (cm *ClusterManager) SetupSharedVolume(ctx context.Context) (string, error) {
	volumeName := cm.volumeName()

	// Check if the volume already exists
	volumes, err := cm.client.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to get Docker volume list: %w", err)
	}

	for _, vol := range volumes.Volumes {
		if vol.Name == volumeName {
			// Volume already exists
			cm.sharedVolumePath = vol.Mountpoint
		}
	}

	// Volume doesn't exist, create it
	vol, err := cm.client.VolumeCreate(
		ctx,
		volume.CreateOptions{
			Driver: "local",
			Name:   volumeName,
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to create Docker volume: %w", err)
	}

	cm.sharedVolumePath = vol.Mountpoint

	return cm.sharedVolumePath, nil
}

// CreateContainer creates a single cluster node container.
// Requires the shared volume and network to be configured beforehand
func (cm *ClusterManager) CreateContainer(ctx context.Context, config ContainerConfig) error {
	// Make sure the shared volume is set up
	if cm.sharedVolumePath == "" {
		return errors.New("cluster shared volume is not created")
	}

	// Make sure the shared network is set up
	if cm.networkID == "" {
		return errors.New("cluster shared network is not created")
	}

	// Create port bindings
	var (
		portBindings = make(nat.PortMap)
		exposedPorts = make(nat.PortSet)
	)

	for containerPort, hostPort := range config.Ports {
		port := nat.Port(containerPort)
		exposedPorts[port] = struct{}{} // mark the port as exposed

		// If host port is empty, it will be auto-assigned
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	// Set up volume mounts for the container.
	// The cluster's shared volume is always part of the
	// mount set
	mounts := []mount.Mount{
		{
			Type:   mount.TypeVolume,
			Source: cm.volumeName(),
			Target: cm.sharedVolumePath,
		},
	}

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
			AutoRemove:   false, // container cleanup is managed by the cluster
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				cm.networkName(): {
					NetworkID: cm.networkID,
				},
			},
		},
		nil,
		config.Name,
	)
	if err != nil {
		return fmt.Errorf("unable to create cluster container: %w", err)
	}

	// Grab the container ID
	containerID := resp.ID

	// Start container.
	// The container needs to be booted in order to get assigned port values.
	// Because of this, it is expected that the containers don't have an "actionable" ENTRYPOINT and CMD,
	// but instead utilize something like "/bin/sh -c sleep infinity" as the entrypoint
	if err := cm.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	// Get container details
	containerJSON, err := cm.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return fmt.Errorf("unable to get cluster container %s details: %w", containerID, err)
	}

	// Get assigned ports
	var (
		containerPorts = containerJSON.NetworkSettings.Ports
		assignedPorts  = make(map[string]string)
	)

	for containerPort, bindings := range containerPorts {
		if len(bindings) == 0 {
			continue
		}

		assignedPorts[string(containerPort)] = bindings[0].HostPort
	}

	// Save container info
	cm.containerInfos[config.Name] = ContainerInfo{
		ID:     containerID,
		Name:   config.Name,
		Image:  config.Image,
		HostIP: "127.0.0.1", // TODO change this? I assume localhost is fine
		Ports:  assignedPorts,
	}

	return nil
}

// ExecuteCmd executes the given command (single) inside the given container.
// Returns the command output, from the container
func (cm *ClusterManager) ExecuteCmd(
	ctx context.Context,
	cName string,
	cmd []string,
) (string, error) {
	// Check if the container exists
	// TODO check
	// cName := cm.containerName(containerName)

	containerInfo, exists := cm.containerInfos[cName]
	if !exists {
		return "", fmt.Errorf("container %s not found", cName)
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
		return "", fmt.Errorf("unable to create exec instance: %w", err)
	}

	// Start exec instance
	resp, err := cm.client.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to attach to exec instance: %w", err)
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
		return outBuf.String(), fmt.Errorf(
			"unable to inspect exec (cmd) instance (container %s): %w",
			cName,
			err,
		)
	}

	if execInspect.ExitCode != 0 {
		return outBuf.String(), fmt.Errorf(
			"container %s command exited with code %d",
			cName,
			execInspect.ExitCode,
		)
	}

	return outBuf.String(), nil
}

// Cleanup stops and removes all cluster containers.
// Additionally, it wipes the shared volumes and networks
func (cm *ClusterManager) Cleanup(ctx context.Context) {
	timeout := 10 // seconds. The Docker API requires an int reference. Don't ask.

	// Gather the container IDs
	containerIDs := make([]string, 0, len(cm.containerInfos))
	for _, info := range cm.containerInfos {
		containerIDs = append(containerIDs, info.ID)
	}

	// Stop the container
	for _, id := range containerIDs {
		// Stop the container
		if err := cm.client.ContainerStop(
			ctx,
			id,
			container.StopOptions{
				Timeout: &timeout,
			},
		); err != nil {
			cm.logger.Error(
				"failed to stop container",
				"id", id,
				"err", err,
			)
		}

		// Remove the container
		if err := cm.client.ContainerRemove(
			ctx,
			id,
			container.RemoveOptions{
				Force: true,
			},
		); err != nil {
			cm.logger.Error(
				"failed to remove container",
				"id", id,
				"err", err,
			)
		}
	}

	// Remove the network, if any
	if cm.networkID != "" {
		if err := cm.client.NetworkRemove(ctx, cm.networkID); err != nil {
			cm.logger.Error(
				"failed to remove network",
				"network ID", cm.networkID,
				"err", err,
			)
		}
	}

	// Remove the shared volume, if any
	if cm.sharedVolumePath != "" {
		volumeName := cm.volumeName()

		if err := cm.client.VolumeRemove(ctx, volumeName, true); err != nil {
			cm.logger.Error(
				"failed to remove volume",
				"volume", volumeName,
				"err", err,
			)
		}
	}

	// Clean the container reference
	clear(cm.containerInfos)
}

// PipeLogs starts capturing logs from the given container and writes them to the given writer.
// Logs are continuously captured, with timestamps
func (cm *ClusterManager) PipeLogs(ctx context.Context, containerName string, w io.WriteCloser) error {
	cName := cm.containerName(containerName)

	// Grab the container info
	containerInfo, exists := cm.containerInfos[cName]
	if !exists {
		return fmt.Errorf("container %s not found", cName)
	}

	// Set the log capture options
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	}

	logs, err := cm.client.ContainerLogs(ctx, containerInfo.ID, opts)
	if err != nil {
		_ = w.Close()

		return fmt.Errorf("unable to get container %s logs: %w", cName, err)
	}

	go func() {
		defer func() {
			// Cleanup
			_ = w.Close()
			_ = logs.Close()
		}()

		// Docker API may include 8-byte headers for each log line.
		// We need to handle this properly. Again, don't ask.
		header := make([]byte, 8)

		for {
			_, err := io.ReadFull(logs, header)
			if err != nil {
				// Check if the log stream is closed (ex. container stopped)
				if !errors.Is(err, io.EOF) {
					cm.logger.Error(
						"error while reading log header",
						"container", containerInfo.ID,
						"err", err,
					)
				}

				// Show-stopping error
				break
			}

			// Get payload size from header (uint32 at position 4)
			payloadSize := int(header[4]) | int(header[5])<<8 | int(header[6])<<16 | int(header[7])<<24

			// Read the payload (log)
			payload := make([]byte, payloadSize)

			_, err = io.ReadFull(logs, payload)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					cm.logger.Error(
						"error while reading log header",
						"container", containerInfo.ID,
						"err", err,
					)
				}

				// Show-stopping error
				break
			}

			if _, err := w.Write(payload); err != nil {
				cm.logger.Error(
					"error while writing container log",
					"container", containerInfo.ID,
					"err", err,
				)

				break
			}
		}
	}()

	return nil
}

// volumeName returns the cluster's shared volume name
// (for shared artifacts like genesis)
func (cm *ClusterManager) volumeName() string {
	return fmt.Sprintf("%s-%s", volumeNamePrefix, cm.name)
}

// networkName returns the cluster's Docker network name
func (cm *ClusterManager) networkName() string {
	return fmt.Sprintf("%s-%s", networkNamePrefix, cm.name)
}

// containerName maps the container name (which can be non-unique)
// to a unique container name (within the cluster)
func (cm *ClusterManager) containerName(name string) string {
	return fmt.Sprintf("%s-container-%s", cm.name, name)
}

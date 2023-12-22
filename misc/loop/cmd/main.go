package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type service struct {
}

func portalLoop(ctx context.Context, s *snapshotter) error {
	logrus.Info("Starting the Portal Loop")

	// 1. Pull latest docker image
	isNew, err := s.pullLatestImage(ctx)
	if err != nil {
		return err
	}

	// 2. Get existing portal loop
	containers, err := s.getPortalLoopContainers(ctx)
	if err != nil {
		return err
	} else if len(containers) == 0 {
		logrus.Info("No portal loop instance found, starting one")
		// Portal loop isn't running, Starting it
		container, err := s.startPortalLoopContainer(context.Background())
		if err != nil {
			return err
		}
		containers = []types.Container{*container}

		for _, p := range container.Ports {
			if p.Type == "tcp" && p.PrivatePort == uint16(26657) {
				s.switchTraefikPortalLoop("http://localhost:" + strconv.Itoa(int(p.PublicPort)))
				break
			}
		}
		return nil
	}

	// 3. Check if there is a new image
	if !isNew {
		// 	return nil
	}

	// 4. Set traefik in READ ONLY mode
	err = s.switchTraefikMode(setReadOnly)
	if err != nil {
		return err
	}

	defer func() {
		err = s.switchTraefikMode(unsetReadOnly)
		if err != nil {
			logrus.WithError(err).Error()
		}
	}()

	// 5. Backup TXs
	err = s.backupTXs(ctx)
	if err != nil {
		return err
	}

	// 6. Start a new portal loop
	container, err := s.startPortalLoopContainer(context.Background())
	if err != nil {
		return err
	}
	for _, p := range container.Ports {
		if p.Type == "tcp" && p.PrivatePort == uint16(26657) {
			s.switchTraefikPortalLoop("http://localhost:" + strconv.Itoa(int(p.PublicPort)))
			break
		}
	}

	// 7. Wait for new portal loop to be ready
	// Wait 5 blocs

	// 8. Remove old portal loop
	for _, c := range containers {
		err = s.dockerClient.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force:         true,  // Force the removal of a running container
			RemoveVolumes: true,  // Remove the volumes associated with the container
			RemoveLinks:   false, // Remove the specified link and not the underlying container
		})
		if err != nil {
			return err
		}
		for _, mount := range c.Mounts {
			if mount.Type == "volume" {
				err = s.dockerClient.VolumeRemove(ctx, mount.Name, true)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func main() {
	var err error

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	for {
		snapshotterClient, err := NewSnapshotter(dockerClient, config{
			backupDir: "./backups",
			rpcAddr:   "http://rpc.portal.gno.local:81",
			hostPWD:   os.Getenv("HOST_PWD"),
		})
		if err != nil {
			logrus.WithError(err).Fatal()
		}

		ctx := context.Background()
		err = portalLoop(ctx, snapshotterClient)
		if err != nil {
			logrus.WithError(err).Error()
		}
		time.Sleep(time.Second * 10)
	}
	// // Rest of the logic continues...
	// // Including waiting for the RPC to be up, getting the RPC port, updating traefik URL
	// // and cleaning up old containers and volumes

	// // Implement logic for the above steps
}

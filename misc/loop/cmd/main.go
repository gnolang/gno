package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func init() {
	if os.Getenv("HOST_PWD") == "" {
		os.Setenv("HOST_PWD", os.Getenv("PWD"))
	}

	if os.Getenv("BACKUP_DIR") == "" {
		os.Setenv("BACKUP_DIR", "./backups")
	}

	if os.Getenv("RPC_URL") == "" {
		os.Setenv("RPC_URL", "http://rpc.portal.gno.local:81")
	}

	if os.Getenv("PROM_ADDR") == "" {
		os.Setenv("PROM_ADDR", ":9090")
	}

	if os.Getenv("TRAEFIK_GNO_FILE") == "" {
		os.Setenv("TRAEFIK_GNO_FILE", "./traefik/gno.yml")
	}
}

type service struct {
	// TODO(albttx): put getter on it with RMutex
	portalLoop *snapshotter

	portalLoopURL string
}

func (s *service) startPortalLoop(ctx context.Context) error {
	logrus.Info("Starting the Portal Loop")

	// 1. Pull latest docker image
	isNew, err := s.portalLoop.pullLatestImage(ctx)
	if err != nil {
		return err
	}

	// 2. Get existing portal loop
	containers, err := s.portalLoop.getPortalLoopContainers(ctx)
	if err != nil {
		return err
	} else if len(containers) == 0 {
		logrus.Info("No portal loop instance found, starting one")
		// Portal loop isn't running, Starting it
		container, err := s.portalLoop.startPortalLoopContainer(context.Background())
		if err != nil {
			return err
		}
		containers = []types.Container{*container}

		for _, p := range container.Ports {
			if p.Type == "tcp" && p.PrivatePort == uint16(26657) {
				s.portalLoopURL = fmt.Sprintf("http://localhost:%d", int(p.PublicPort))
				s.portalLoop.switchTraefikPortalLoop(s.portalLoopURL)
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
	err = s.portalLoop.switchTraefikMode(setReadOnly)
	if err != nil {
		return err
	}

	defer func() {
		err = s.portalLoop.switchTraefikMode(unsetReadOnly)
		if err != nil {
			logrus.WithError(err).Error()
		}
	}()

	// 5. Backup TXs
	err = s.portalLoop.backupTXs(ctx)
	if err != nil {
		return err
	}

	// 6. Start a new portal loop
	container, err := s.portalLoop.startPortalLoopContainer(context.Background())
	if err != nil {
		return err
	}
	for _, p := range container.Ports {
		if p.Type == "tcp" && p.PrivatePort == uint16(26657) {
			s.portalLoopURL = fmt.Sprintf("http://localhost:%d", int(p.PublicPort))
			s.portalLoop.switchTraefikPortalLoop("http://localhost:" + strconv.Itoa(int(p.PublicPort)))
			break
		}
	}

	// 7. Wait for new portal loop to be ready
	// Wait 5 blocs

	// 8. Remove old portal loop
	for _, c := range containers {
		err = s.portalLoop.dockerClient.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force:         true,  // Force the removal of a running container
			RemoveVolumes: true,  // Remove the volumes associated with the container
			RemoveLinks:   false, // Remove the specified link and not the underlying container
		})
		if err != nil {
			return err
		}
		for _, mount := range c.Mounts {
			if mount.Type == "volume" {
				err = s.portalLoop.dockerClient.VolumeRemove(ctx, mount.Name, true)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *service) ServeMonitoring() {
	// Wait for portal loop to start
	for s.portalLoopURL == "" {
		time.Sleep(time.Second * 1)
	}

	go s.recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(os.Getenv("PROM_ADDR"), nil)
}

func main() {
	s := &service{}

	var err error

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	go s.ServeMonitoring()

	for {
		s.portalLoop, err = NewSnapshotter(dockerClient, config{
			backupDir:      os.Getenv("BACKUP_DIR"),
			rpcAddr:        os.Getenv("RPC_URL"),
			hostPWD:        os.Getenv("HOST_PWD"),
			traefikGnoFile: os.Getenv("TRAEFIK_GNO_FILE"),
		})
		if err != nil {
			logrus.WithError(err).Fatal()
		}

		ctx := context.Background()
		err = s.startPortalLoop(ctx)
		if err != nil {
			logrus.WithError(err).Error()
		}
		time.Sleep(time.Second * 10)
	}
}

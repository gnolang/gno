package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	ff "github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

type service struct {
	// TODO(albttx): put getter on it with RMutex
	portalLoop *snapshotter

	portalLoopURL string
}

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

func main() {
	s := &service{}

	serveCmd := &ff.Command{
		Name: "serve",
		Exec: s.execServe,
	}
	switchCmd := &ff.Command{
		Name: "switch",
		Exec: s.execSwitch,
	}
	backupCmd := &ff.Command{
		Name: "backup",
		Exec: s.execBackup,
	}

	rootCmd := &ff.Command{
		Name: "portalloopd",
		Subcommands: []*ff.Command{
			serveCmd,
			switchCmd,
			backupCmd,
		},
	}

	if err := rootCmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Command(rootCmd))
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

}

func (s *service) execServe(ctx context.Context, args []string) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	// Serve monitoring
	go func() {
		for s.portalLoopURL == "" {
			time.Sleep(time.Second * 1)
		}

		go s.recordMetrics()

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(os.Getenv("PROM_ADDR"), nil)

	}()

	// the loop
	for {
		s.portalLoop, err = NewSnapshotter(dockerClient, config{
			backupDir:      os.Getenv("BACKUP_DIR"),
			rpcAddr:        os.Getenv("RPC_URL"),
			hostPWD:        os.Getenv("HOST_PWD"),
			traefikGnoFile: os.Getenv("TRAEFIK_GNO_FILE"),
		})
		if err != nil {
			return err
		}

		err = s.startPortalLoop(ctx, false)
		if err != nil {
			logrus.WithError(err).Error()
		}
		time.Sleep(time.Second * 10)
	}
}

func (s *service) execSwitch(ctx context.Context, args []string) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	s.portalLoop, err = NewSnapshotter(dockerClient, config{
		backupDir:      os.Getenv("BACKUP_DIR"),
		rpcAddr:        os.Getenv("RPC_URL"),
		hostPWD:        os.Getenv("HOST_PWD"),
		traefikGnoFile: os.Getenv("TRAEFIK_GNO_FILE"),
	})
	if err != nil {
		return err
	}

	return s.startPortalLoop(ctx, true)
}

func (s *service) execBackup(ctx context.Context, args []string) error {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	s.portalLoop, err = NewSnapshotter(dockerClient, config{
		backupDir:      os.Getenv("BACKUP_DIR"),
		rpcAddr:        os.Getenv("RPC_URL"),
		hostPWD:        os.Getenv("HOST_PWD"),
		traefikGnoFile: os.Getenv("TRAEFIK_GNO_FILE"),
	})
	if err != nil {
		return err
	}

	err = s.startPortalLoop(ctx, false)
	if err != nil {
		return err
	}

	return s.portalLoop.backupTXs(ctx)
}

func (s *service) startPortalLoop(ctx context.Context, force bool) error {
	logrus.Info("Starting the Portal Loop")

	// 1. Pull latest docker image
	isNew, err := s.portalLoop.pullLatestImage(ctx)
	if err != nil {
		return err
	}
	logrus.WithField("is_new", isNew).Info("Pulled latest image")

	// 2. Get existing portal loop
	containers, err := s.portalLoop.getPortalLoopContainers(ctx)
	if err != nil {
		return err
	}
	logrus.WithField("containers", containers).Info("Get containers")

	if len(containers) == 0 {
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

	for _, p := range containers[0].Ports {
		if p.Type == "tcp" && p.PrivatePort == uint16(26657) {
			s.portalLoopURL = fmt.Sprintf("http://localhost:%d", int(p.PublicPort))
			s.portalLoop.switchTraefikPortalLoop(s.portalLoopURL)
			break
		}
	}

	logrus.Info("Current Portal Loop is running on : ", s.portalLoopURL)

	// 3. Check if there is a new image
	if !isNew && !force {
		return nil
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
			break
		}
	}

	// 7. Wait 5 blocks new portal loop to be ready
	now := time.Now()
	for {
		if time.Since(now) > time.Second*120 {
			return fmt.Errorf("timeout getting latest block")
		}
		err := func() error {
			resp, err := http.Get(s.portalLoopURL + "/status")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			tmStatus := struct {
				Result struct {
					SyncInfo struct {
						LatestBlockHeight string `json:"latest_block_height"`
					} `json:"sync_info"`
				} `json:"result"`
			}{}
			if err := json.NewDecoder(resp.Body).Decode(&tmStatus); err != nil {
				return err
			}

			currentBlock, err := strconv.Atoi(tmStatus.Result.SyncInfo.LatestBlockHeight)
			if err != nil {
				return err
			}
			if currentBlock >= 5 {
				return nil
			}
			return fmt.Errorf("blocks: %d/5", currentBlock)
		}()
		if err == nil {
			break
		}

		if !strings.HasPrefix(err.Error(), "blocks: ") {
			logrus.WithError(err).Error()
		}
		time.Sleep(time.Second * 2)
	}

	// 8. Update traefik portal loop rpc url
	if err := s.portalLoop.switchTraefikPortalLoop(s.portalLoopURL); err != nil {
		return err
	}

	// 9. Remove old portal loop
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

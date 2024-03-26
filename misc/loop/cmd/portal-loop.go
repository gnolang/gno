package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
)

func StartPortalLoop(ctx context.Context, portalLoop *snapshotter, force bool) error {
	l := logrus.WithFields(logrus.Fields{})

	// 1. Pull latest docker image
	isNew, err := portalLoop.pullLatestImage(ctx)
	if err != nil {
		return err
	}
	l.WithField("is_new", isNew).Info("Starting the Portal Loop")

	// 2. Get existing portal loop
	containers, err := portalLoop.getPortalLoopContainers(ctx)
	if err != nil {
		return err
	}
	l.WithField("containers", containers).Info("Get containers")

	if len(containers) == 0 {
		logrus.Info("No portal loop instance found, starting one")
		// Portal loop isn't running, Starting it
		container, err := portalLoop.startPortalLoopContainer(context.Background())
		if err != nil {
			return err
		}
		containers = []types.Container{*container}
		force = true
	}

	for _, p := range containers[0].Ports {
		if p.Type == "tcp" && p.PrivatePort == uint16(26657) {
			ip := containers[0].NetworkSettings.Networks["portal-loop"].IPAddress
			portalLoop.url = fmt.Sprintf("http://%s:%d", ip, int(p.PrivatePort))
			portalLoop.switchTraefikPortalLoop(portalLoop.url)
			break
		}
	}

	l = l.WithFields(logrus.Fields{
		"portal.url": portalLoop.url,
	})
	l.Info("Current portal loop")

	// 3. Check if there is a new image
	if !isNew && !force {
		return nil
	}

	l.Info("Set read only mode")
	// 4. Set traefik in READ ONLY mode
	err = portalLoop.switchTraefikMode(setReadOnly)
	if err != nil {
		return err
	}

	defer func() {
		l.Info("Unset read only mode")
		err = portalLoop.switchTraefikMode(unsetReadOnly)
		if err != nil {
			logrus.WithError(err).Error()
		}
	}()

	l.Info("Backup txs")
	// 5. Backup TXs
	err = portalLoop.backupTXs(ctx, portalLoop.url)
	if err != nil {
		return err
	}

	// 6. Start a new portal loop
	container, err := portalLoop.startPortalLoopContainer(context.Background())
	if err != nil {
		return err
	}
	for _, p := range container.Ports {
		if p.Type == "tcp" && p.PrivatePort == uint16(26657) {
			ip := container.NetworkSettings.Networks["portal-loop"].IPAddress
			portalLoop.url = fmt.Sprintf("http://%s:%d", ip, int(p.PrivatePort))
			break
		}
	}
	l = l.WithFields(logrus.Fields{
		"new_portal.url": portalLoop.url,
	})
	l.Info("setup new portal loop")

	// 7. Wait 5 blocks new portal loop to be ready
	now := time.Now()
	for {
		if time.Since(now) > time.Second*120 {
			return fmt.Errorf("timeout getting latest block")
		}
		err := func() error {
			resp, err := http.Get(portalLoop.url + "/status")
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
			l.WithField("new_portal.current_block", currentBlock)

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

	l.Info("update traefik portal loop url")
	// 8. Update traefik portal loop rpc url
	if err := portalLoop.switchTraefikPortalLoop(portalLoop.url); err != nil {
		return err
	}

	// 9. Remove old portal loop
	for _, c := range containers {
		l.WithFields(logrus.Fields{
			"container.id":    c.ID,
			"container.ports": c.Ports,
		}).Infof("remove container")
		err = portalLoop.dockerClient.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force:         true,  // Force the removal of a running container
			RemoveVolumes: true,  // Remove the volumes associated with the container
			RemoveLinks:   false, // Remove the specified link and not the underlying container
		})
		if err != nil {
			return err
		}
		for _, mount := range c.Mounts {
			if mount.Type == "volume" {
				err = portalLoop.dockerClient.VolumeRemove(ctx, mount.Name, true)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

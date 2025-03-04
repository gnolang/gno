package portalloop

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

func StartPortalLoop(ctx context.Context, portalLoopHandler PortalLoopHandler, force bool) error {
	l := logrus.WithFields(logrus.Fields{})
	dockerHandler := portalLoopHandler.dockerHandler

	// 1. Pull latest docker image
	isNew, err := dockerHandler.NeedsPullNewMasterImage(ctx)
	if err != nil {
		return err
	}
	l.WithField("is_new", isNew).Info("Starting the Portal Loop")

	// 2. Get existing portal loop
	containers, err := dockerHandler.GetActiveGnoPortalLoopContainers(ctx)
	if err != nil {
		return err
	}
	l.WithField("containers", containers).Info("Get containers")

	if len(containers) == 0 {
		logrus.Info("No portal loop instance found, starting one")
		// Portal loop isn't running, Starting it
		container, err := dockerHandler.StartGnoPortalLoopContainer(
			ctx,
			portalLoopHandler.containerName,
			portalLoopHandler.cfg.HostPWD,
		)
		if err != nil {
			return err
		}
		containers = []types.Container{*container}
		force = true
	}

	portalLoopHandler.currentRpcUrl = dockerHandler.GetPublishedRPCPort(containers[0])
	portalLoopHandler.SwitchTraefikPortalLoopUrl()

	l.Info("Current portal loop container:")
	l = l.WithFields(logrus.Fields{"portal.url": portalLoopHandler.currentRpcUrl})

	// 3. Check if there is a new image
	if !isNew && !force {
		return nil
	}

	l.Info("Set read only mode")
	// 4. Set traefik in READ ONLY mode
	err = portalLoopHandler.SwitchTraefikMode(true)
	if err != nil {
		return err
	}
	defer func() {
		l.Info("Unset read only mode")
		err = portalLoopHandler.SwitchTraefikMode(false)
		if err != nil {
			logrus.WithError(err).Error()
		}
	}()

	l.Info("Backup txs")
	// 5. Backup TXs
	err = portalLoopHandler.BackupTXs(ctx)
	if err != nil {
		return err
	}

	// 6. Start a new portal loop
	dockerContainer, err := dockerHandler.StartGnoPortalLoopContainer(
		ctx,
		portalLoopHandler.containerName,
		portalLoopHandler.cfg.HostPWD,
	)
	if err != nil {
		return err
	}

	portalLoopHandler.currentRpcUrl = dockerHandler.GetPublishedRPCPort(*dockerContainer)
	l = l.WithFields(logrus.Fields{
		"new_portal.url": portalLoopHandler.currentRpcUrl,
	})
	l.Info("setup new portal loop")

	// 7. Wait 5 blocks new portal loop to be ready
	err = waitStartedLoop(portalLoopHandler.currentRpcUrl)
	if err != nil {
		return err
	}

	l.Info("update traefik portal loop url")
	// 8. Update traefik portal loop rpc url
	if err := portalLoopHandler.SwitchTraefikPortalLoopUrl(); err != nil {
		return err
	}

	// 9. Remove old portal loop --- Should be performed by WatchTower
	return dockerHandler.RemoveContainersWithVolumes(ctx, containers)
}

// Waits for the Loop to get started
func waitStartedLoop(url string) error {
	l := logrus.WithFields(logrus.Fields{})
	now := time.Now()

	for {
		if time.Since(now) > time.Second*120 {
			return fmt.Errorf("timeout getting latest block")
		}
		err := func() error {
			resp, err := http.Get(url + "/status")
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
	return nil
}

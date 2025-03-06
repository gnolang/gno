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

	"go.uber.org/zap"
)

// Runs a Portal Loop routine
func RunPortalLoop(ctx context.Context, portalLoopHandler PortalLoopHandler, force bool) error {
	logger := portalLoopHandler.logger
	dockerHandler := portalLoopHandler.dockerHandler

	// 1. Pull latest docker image
	isNew, err := dockerHandler.CheckPulledMasterImage(ctx)
	if err != nil {
		return err
	}
	logger.Info("Starting the Portal Loop",
		zap.Bool("is_new", isNew),
		zap.Bool("is_forced", force),
	)

	// - Backup existing txs using an existing Gno container / create new one
	// - Kill the existing container
	// - Create a new instance of instance of Gno container that start using backup as txs in genesis file
	// - Wait genesis txs to be committed
	// - Update reference to looped RPC

	// 2. Get existing portal loop
	containers, err := dockerHandler.GetActiveGnoPortalLoopContainers(ctx)
	if err != nil {
		return err
	}
	logger.Info("Get containers",
		zap.Reflect("container", containers),
	)

	if len(containers) == 0 {
		logger.Info("No portal loop instance found, starting one")
		// Portal loop isn't running, Starting it
		container, err := dockerHandler.StartGnoPortalLoopContainer(
			ctx,
			portalLoopHandler.containerName,
			portalLoopHandler.cfg.HostPWD,
			false, // do not pull new image
		)
		if err != nil {
			return err
		}
		containers = []types.Container{*container}
		force = true // force performing all the steps
	}

	portalLoopHandler.currentRpcUrl = dockerHandler.GetPublishedRPCPort(containers[0])
	portalLoopHandler.SwitchTraefikPortalLoopUrl()
	logger.Info("Current portal loop container",
		zap.String("portal.url", portalLoopHandler.currentRpcUrl),
	)

	// 3. Check image or options. DO not proceed, if not any new docker image AND not forced loop
	if !isNew && !force {
		return nil
	}

	// 4. Set Traefik in READ ONLY mode
	logger.Info("Setting read only mode")
	err = portalLoopHandler.LockTraefikAccess()
	if err != nil {
		return err
	}
	defer func() {
		logger.Info("Unsetting read only mode")
		err = portalLoopHandler.UnlockTraefikAccess()
		if err != nil {
			logrus.WithError(err).Error()
		}
	}()

	// 5. Backup TXs
	logger.Info("Backup txs")
	err = portalLoopHandler.BackupTXs(ctx)
	if err != nil {
		return err
	}

	// 6. Start a new portal loop instance
	dockerContainer, err := dockerHandler.StartGnoPortalLoopContainer(
		ctx,
		portalLoopHandler.containerName,
		portalLoopHandler.cfg.HostPWD,
		true, // always pull new image
	)
	if err != nil {
		return err
	}
	portalLoopHandler.currentRpcUrl = dockerHandler.GetPublishedRPCPort(*dockerContainer)
	logger.Info("Set up new portal loop container",
		zap.String("portal.url", portalLoopHandler.currentRpcUrl),
	)

	// 7. Wait first blocks meaning new portal loop to be ready
	// Backups will be restored from genesis file
	if err = waitStartedLoop(portalLoopHandler.currentRpcUrl, portalLoopHandler.logger); err != nil {
		return err
	}

	// 8. Update traefik portal loop rpc url
	logger.Info("Updating Traefik portal loop url")
	if err = portalLoopHandler.SwitchTraefikPortalLoopUrl(); err != nil {
		return err
	}

	// 9. Remove old portal loop
	return dockerHandler.RemoveContainersWithVolumes(ctx, containers)
}

// Waits for the Loop to get started
func waitStartedLoop(url string, logger *zap.Logger) error {
	now := time.Now()
	for {
		if time.Since(now) > time.Second*180 {
			return fmt.Errorf("timeout getting latest block")
		}
		err := checkCurrentBlock(url)
		if err == nil {
			logger.Info("Loop is finished")
			break
		}

		if !strings.HasPrefix(err.Error(), "blocks: ") {
			logger.Error("Error fetching blocks", zap.Error(err))
		}
		time.Sleep(time.Second * 2)
	}
	return nil
}

// Gets Current Block from /status endpoint in RPC node
func checkCurrentBlock(url string) error {
	resp, err := http.Get(url + "/status")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmStatus := TendermintStatus{}
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
}

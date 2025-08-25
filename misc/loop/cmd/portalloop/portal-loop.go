package portalloop

import (
	"context"
	"log/slog"

	"github.com/docker/docker/api/types"
)

// Runs a Portal Loop routine
// Steps performed:
// - Check for an updated Gnoland Docker image (if not skip)
// - Backup existing txs using an existing Gnoland container / create new one
// - Kill the existing container
// - Create a new instance of Gnoland container that starts using the previous backup as txs in its genesis file
// - Wait genesis txs to be committed
// - Update reference to looped RPC node
func RunPortalLoop(ctx context.Context, portalLoopHandler PortalLoopHandler, force bool) error {
	logger := portalLoopHandler.logger
	dockerHandler := portalLoopHandler.dockerHandler

	// 1. Pull latest docker image
	isNew, err := dockerHandler.CheckPulledMasterImage(ctx)
	if err != nil {
		return err
	}
	logger.Info("Starting the Portal Loop",
		slog.Bool("_new", isNew),
		slog.Bool("_forced", force),
	)

	// 2. Get existing portal loop
	containers, err := dockerHandler.GetActiveGnoPortalLoopContainers(ctx)
	if err != nil {
		return err
	}
	logger.Info("Get containers",
		slog.Any("containers", containers),
	)

	if len(containers) == 0 {
		logger.Info("No portal loop instance found, starting one")
		// Portal loop isn't running, Starting it
		container, err := dockerHandler.StartGnoPortalLoopContainer(
			ctx,
			portalLoopHandler.GetContainerName(),
			portalLoopHandler.cfg.HostPwd,
			false, // do not pull new image
		)
		if err != nil {
			return err
		}
		containers = []types.Container{*container}
		force = true // force performing all the steps
	}

	portalLoopHandler.currentRpcUrl = dockerHandler.GetPublishedRPCPort(containers[0])
	logger.Info("Current loop container",
		slog.String("portal.url", portalLoopHandler.currentRpcUrl),
	)
	logger.Info("Updating Traefik portal loop url")
	if err := portalLoopHandler.UpdateTraefikPortalLoopUrl(); err != nil {
		return err
	}

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
			logger.Error("Issues whil unsetting read-only mode", slog.Any("err", err))
		}
	}()

	// 5. Backup TXs
	if err = portalLoopHandler.WaitStartedLoop(); err != nil {
		return err
	}
	logger.Info("Backup txs")
	err = portalLoopHandler.BackupTXs(ctx)
	if err != nil {
		return err
	}

	// 6. Start a new portal loop instance
	dockerContainer, err := dockerHandler.StartGnoPortalLoopContainer(
		ctx,
		portalLoopHandler.GetContainerName(),
		portalLoopHandler.cfg.HostPwd,
		true, // always pull new image
	)
	if err != nil {
		return err
	}
	portalLoopHandler.currentRpcUrl = dockerHandler.GetPublishedRPCPort(*dockerContainer)
	logger.Info("Updated portal loop container",
		slog.String("portal.url", portalLoopHandler.currentRpcUrl),
	)

	// 7. Wait first blocks meaning new portal loop to be ready
	// Backups will be restored from genesis file
	if err = portalLoopHandler.WaitStartedLoop(); err != nil {
		return err
	}

	// 8. Update traefik portal loop rpc url
	logger.Info("Updating Traefik portal loop url")
	if err = portalLoopHandler.UpdateTraefikPortalLoopUrl(); err != nil {
		return err
	}

	// 9. Remove old portal loop
	return dockerHandler.RemoveContainersWithVolumes(ctx, containers)
}

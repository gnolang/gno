package portalloop

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/misc/loop/cmd/cfg"
	"github.com/gnolang/gno/misc/loop/cmd/docker"

	dockerClient "github.com/docker/docker/client"
	"github.com/gnolang/gno/contribs/tx-archive/backup"
	"github.com/gnolang/gno/contribs/tx-archive/backup/client/rpc"
	"github.com/gnolang/gno/contribs/tx-archive/backup/writer/standard"
)

type TraefikMode string

const (
	setReadOnly   TraefikMode = `middlewares: ["ipwhitelist"]`
	unsetReadOnly TraefikMode = `middlewares: []`
)

type PortalLoopHandler struct {
	cfg                *cfg.CmdCfg
	logger             *slog.Logger
	dockerHandler      *docker.DockerHandler
	currentRpcUrl      string
	backupFile         string
	instanceBackupFile string
}

// Gets formatted current time
func getFormattedTimestamp() string {
	timenow := time.Now()
	return timenow.Format("2006-01-02_") + strconv.FormatInt(timenow.UnixNano(), 10)
}

func NewPortalLoopHandler(cfg *cfg.CmdCfg, logger *slog.Logger) (*PortalLoopHandler, error) {
	dockerClient_, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return nil, err
	}

	// The master backup file will contain the ultimate txs backup
	// that the portal loop use when looping, generating the genesis.
	backupFile, err := filepath.Abs(cfg.MasterBackupFile)
	if err != nil {
		return nil, err
	}
	instanceBackupFile, err := filepath.Abs(fmt.Sprintf("%s/backup_%s.jsonl", cfg.SnapshotsDir, getFormattedTimestamp()))
	if err != nil {
		return nil, err
	}

	return &PortalLoopHandler{
		cfg:    cfg,
		logger: logger,
		dockerHandler: &docker.DockerHandler{
			DockerClient: dockerClient_,
			Logger:       logger,
		},
		backupFile:         backupFile,
		instanceBackupFile: instanceBackupFile,
	}, nil
}

// Gets a container name from current time
func (plh PortalLoopHandler) GetContainerName() string {
	return "gno" + getFormattedTimestamp()
}

// Backups all the active transactions
func (plh PortalLoopHandler) BackupTXs(ctx context.Context) error {
	cfg := backup.DefaultConfig()
	cfg.FromBlock = 1
	cfg.Watch = false

	// We want to skip failed txs on the Portal Loop reset,
	// because they might (unexpectedly) succeed
	cfg.SkipFailedTx = true

	instanceBackupFile, err := os.Create(plh.instanceBackupFile)
	if err != nil {
		return err
	}
	defer instanceBackupFile.Close()

	w := standard.NewWriter(instanceBackupFile)

	// Create the tx-archive backup service
	c, err := rpc.NewHTTPClient(plh.currentRpcUrl)
	if err != nil {
		return fmt.Errorf("could not create tx-archive client, %w", err)
	}

	backupService := backup.NewService(c, w)

	// Run the backup service
	if backupErr := backupService.ExecuteBackup(ctx, cfg); backupErr != nil {
		return fmt.Errorf("unable to execute backup, %w", backupErr)
	}

	if err := instanceBackupFile.Sync(); err != nil {
		return err
	}

	info, err := instanceBackupFile.Stat()
	if err != nil {
		return err
	} else if info.Size() == 0 {
		return os.Remove(instanceBackupFile.Name())
	}

	// Append to backup file
	backupFile, err := os.OpenFile(plh.backupFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file %s, %w", plh.backupFile, err)
	}
	defer backupFile.Close()

	output, err := os.ReadFile(plh.instanceBackupFile)
	if err != nil {
		return err
	}

	_, err = backupFile.Write(output)
	return err
}

// Replaces string mathcing Regexp in a given file
func (plh *PortalLoopHandler) replaceRegexpInFile(regExp string, replaceStr string) error {
	input, err := os.ReadFile(plh.cfg.TraefikGnoFile)
	if err != nil {
		return err
	}

	regex := regexp.MustCompile(regExp)
	output := regex.ReplaceAllLiteral(input, []byte(replaceStr))

	return os.WriteFile(plh.cfg.TraefikGnoFile, output, 0o655)
}

// Replaces URL of Gno Portal Loop
func (plh *PortalLoopHandler) UpdateTraefikPortalLoopUrl() error {
	regExp := `http://.*:[0-9]+`
	return plh.replaceRegexpInFile(regExp, plh.currentRpcUrl)
}

// Replaces Traefik Mode attribute - set/unset ReadOnly mode
func (plh *PortalLoopHandler) switchTraefikAccess(mode TraefikMode) error {
	regExp := `middlewares: \[.*\]`
	return plh.replaceRegexpInFile(regExp, (string)(mode))
}

// Unsets ReadOnly mode
func (plh *PortalLoopHandler) UnlockTraefikAccess() error {
	return plh.switchTraefikAccess(unsetReadOnly)
}

// Sets ReadOnly mode
func (plh *PortalLoopHandler) LockTraefikAccess() error {
	return plh.switchTraefikAccess(setReadOnly)
}

// Proxies the request to the Docker handler to remove existing Portal Lopp containers
func (plh *PortalLoopHandler) ProxyRemoveContainers(ctx context.Context) error {
	containers, err := plh.dockerHandler.GetActiveGnoPortalLoopContainers(ctx)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return nil
	}
	return plh.dockerHandler.RemoveContainersWithVolumes(ctx, containers)
}

// Waits for the Loop to get started
func (plh *PortalLoopHandler) WaitStartedLoop() error {
	done := make(chan struct{})

	go func() {
		for {
			err := plh.checkCurrentBlock()
			if err == nil {
				close(done)
				break
			}

			if strings.HasPrefix(err.Error(), "blocks: ") {
				plh.logger.Error("Fetched blocks", slog.Any("err", err))
			} else {
				plh.logger.Error("Error fetching blocks", slog.Any("err", err))
			}
			time.Sleep(5 * time.Second)
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Minute):
		return fmt.Errorf("timeout getting latest block")
	}
}

// Gets Current Block from /status endpoint in RPC node
func (plh *PortalLoopHandler) checkCurrentBlock() error {
	resp, err := http.Get(plh.currentRpcUrl + "/status")
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
	return fmt.Errorf("%d/5", currentBlock)
}

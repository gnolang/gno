package portalloop

import (
	"context"
	"fmt"
	"loop/cmd/cfg"
	"loop/cmd/docker"
	"os"
	"path/filepath"
	"regexp"
	"time"

	dockerClient "github.com/docker/docker/client"
	"github.com/gnolang/tx-archive/backup"
	"github.com/gnolang/tx-archive/backup/client/rpc"
	"github.com/gnolang/tx-archive/backup/writer/standard"
	"go.uber.org/zap"
)

type TraefikMode string

const (
	setReadOnly   TraefikMode = `middlewares: ["ipwhitelist"]`
	unsetReadOnly TraefikMode = `middlewares: []`
)

type PortalLoopHandler struct {
	cfg           *cfg.CmdCfg
	logger        *zap.Logger
	dockerHandler *docker.DockerHandler
	containerName string
	currentRpcUrl string

	backupFile         string
	instanceBackupFile string
}

func NewPortalLoopHandler(cfg *cfg.CmdCfg, logger *zap.Logger) (*PortalLoopHandler, error) {
	timenow := time.Now()
	now := fmt.Sprintf("%s_%v", timenow.Format("2006-01-02_"), timenow.UnixNano())

	dockerClient_, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		return nil, err
	}

	backupFile, err := filepath.Abs(cfg.MasterBackupFile)
	if err != nil {
		return nil, err
	}
	instanceBackupFile, err := filepath.Abs(fmt.Sprintf("%s/backup_%s.jsonl", cfg.SnapshotsDir, now))
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
		containerName:      "gno-" + now,
		backupFile:         backupFile,
		instanceBackupFile: instanceBackupFile,
	}, nil
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
func (plh *PortalLoopHandler) SwitchTraefikPortalLoopUrl() error {
	regExp := `http://.*:[0-9]+`
	return plh.replaceRegexpInFile(regExp, plh.currentRpcUrl)
}

// Replaces Traefik Mode attribute - set/unset ReadOnly mode
func (plh *PortalLoopHandler) SwitchTraefikMode(readOnly bool) error {
	regExp := `middlewares: \[.*\]`

	var mode TraefikMode = unsetReadOnly
	if readOnly {
		mode = setReadOnly
	}
	return plh.replaceRegexpInFile(regExp, (string)(mode))
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

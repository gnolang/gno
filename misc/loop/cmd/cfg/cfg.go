package cfg

import (
	"flag"
	"os"
)

const (
	HOST_PWD                = "HOST_PWD"
	TRAEFIK_GNO_FILE        = "TRAEFIK_GNO_FILE"
	MASTER_BACKUP_FILE      = "MASTER_BACKUP_FILE"
	SNAPSHOTS_DIR           = "SNAPSHOTS_DIR"
	defaultTraefiKPath      = "./traefik/gno.yml"
	defaultMasterBackupFile = "./backups/backup.jsonl"
	defaultSnapshotFolder   = "./backups/snapshots"
)

type CmdCfg struct {
	HostPwd          string
	TraefikGnoFile   string
	MasterBackupFile string
	SnapshotsDir     string
}

func (cfg *CmdCfg) RegisterFlags(fs *flag.FlagSet) {
	if os.Getenv(HOST_PWD) == "" {
		os.Setenv(HOST_PWD, os.Getenv("PWD"))
	}

	if os.Getenv(TRAEFIK_GNO_FILE) == "" {
		os.Setenv(TRAEFIK_GNO_FILE, defaultTraefiKPath)
	}

	if os.Getenv(MASTER_BACKUP_FILE) == "" {
		os.Setenv(MASTER_BACKUP_FILE, defaultMasterBackupFile)
	}

	if os.Getenv(SNAPSHOTS_DIR) == "" {
		os.Setenv(SNAPSHOTS_DIR, defaultSnapshotFolder)
	}

	fs.StringVar(&cfg.HostPwd, "pwd", os.Getenv("HOST_PWD"), "host pwd (for docker usage)")
	fs.StringVar(&cfg.TraefikGnoFile, "traefik-gno-file", os.Getenv("TRAEFIK_GNO_FILE"), "traefik gno file")
	fs.StringVar(&cfg.MasterBackupFile, "master-backup-file", os.Getenv("MASTER_BACKUP_FILE"), "master txs backup file path")
	fs.StringVar(&cfg.SnapshotsDir, "snapshots-dir", os.Getenv("SNAPSHOTS_DIR"), "snapshots directory")
}

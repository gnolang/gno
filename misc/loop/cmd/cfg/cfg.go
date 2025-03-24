package cfg

import (
	"flag"
	"os"
)

type CmdCfg struct {
	RpcAddr        string
	TraefikGnoFile string
	HostPWD        string

	MasterBackupFile string
	SnapshotsDir     string
}

func (cfg *CmdCfg) RegisterFlags(fs *flag.FlagSet) {

	if os.Getenv("HOST_PWD") == "" {
		os.Setenv("HOST_PWD", os.Getenv("PWD"))
	}

	if os.Getenv("SNAPSHOTS_DIR") == "" {
		os.Setenv("SNAPSHOTS_DIR", "./backups/snapshots")
	}

	if os.Getenv("RPC_URL") == "" {
		os.Setenv("RPC_URL", "http://rpc.portal.gno.local:26657")
	}

	if os.Getenv("PROM_ADDR") == "" {
		os.Setenv("PROM_ADDR", ":9090")
	}

	if os.Getenv("TRAEFIK_GNO_FILE") == "" {
		os.Setenv("TRAEFIK_GNO_FILE", "./traefik/gno.yml")
	}

	if os.Getenv("MASTER_BACKUP_FILE") == "" {
		os.Setenv("MASTER_BACKUP_FILE", "./backups/backup.jsonl")
	}

	fs.StringVar(&cfg.RpcAddr, "rpc", os.Getenv("RPC_URL"), "tendermint rpc url")
	fs.StringVar(&cfg.TraefikGnoFile, "traefik-gno-file", os.Getenv("TRAEFIK_GNO_FILE"), "traefik gno file")
	fs.StringVar(&cfg.HostPWD, "pwd", os.Getenv("HOST_PWD"), "host pwd (for docker usage)")
	fs.StringVar(&cfg.MasterBackupFile, "master-backup-file", os.Getenv("MASTER_BACKUP_FILE"), "master txs backup file path")
	fs.StringVar(&cfg.SnapshotsDir, "snapshots-dir", os.Getenv("SNAPSHOTS_DIR"), "snapshots directory")
}

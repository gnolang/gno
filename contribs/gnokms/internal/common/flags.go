package common

import (
	"flag"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"go.uber.org/zap/zapcore"
)

type Flags struct {
	ChainID           string
	NodeAddr          string
	DialTimeout       time.Duration
	DialMaxRetries    uint
	DialRetryInterval time.Duration
	ReadWriteTimeout  time.Duration
	LogLevel          string
	LogFormat         string
}

var defaultFlags = Flags{
	ChainID:           "dev",
	NodeAddr:          "tcp://127.0.0.1:26659",
	DialTimeout:       time.Second * 3,
	DialMaxRetries:    privval.DefaultMaxDialRetries,
	DialRetryInterval: privval.DefaultDialRetryIntervalMS * time.Millisecond,
	ReadWriteTimeout:  privval.DefaultReadWriteTimeoutSeconds * time.Second,
	LogLevel:          zapcore.DebugLevel.String(),
	LogFormat:         log.ConsoleFormat.String(),
}

func (f *Flags) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&f.ChainID,
		"chainid",
		defaultFlags.ChainID,
		"the ID of the chain",
	)

	fs.StringVar(
		&f.NodeAddr,
		"node-addr",
		defaultFlags.NodeAddr,
		"TCP or UNIX address of the node",
	)

	fs.DurationVar(
		&f.DialTimeout,
		"tcp-timeout",
		defaultFlags.DialTimeout,
		"timeout for dialing node using TCP",
	)

	fs.UintVar(
		&f.DialMaxRetries,
		"max-retries",
		defaultFlags.DialMaxRetries,
		"maximum number of retries to dial node",
	)

	fs.DurationVar(
		&f.DialRetryInterval,
		"retry-interval",
		defaultFlags.DialRetryInterval,
		"interval to wait between dial retries",
	)

	fs.DurationVar(
		&f.ReadWriteTimeout,
		"rw-timeout",
		defaultFlags.ReadWriteTimeout,
		"timeout for read/write operations",
	)

	fs.StringVar(
		&f.LogLevel,
		"log-level",
		defaultFlags.LogLevel,
		"log level (debug|info|warn|error)",
	)

	fs.StringVar(
		&f.LogFormat,
		"log-format",
		defaultFlags.LogFormat,
		"log format (json|console)",
	)
}

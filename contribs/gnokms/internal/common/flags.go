package common

import (
	"flag"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/log"
	sserver "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
	"go.uber.org/zap/zapcore"
)

type Flags struct {
	ListenAddresses string
	KeepAlivePeriod time.Duration
	ResponseTimeout time.Duration
	LogLevel        string
	LogFormat       string
}

var defaultFlags = Flags{
	ListenAddresses: "tcp://127.0.0.1:26659",
	KeepAlivePeriod: sserver.DefaultKeepAlivePeriod,
	ResponseTimeout: sserver.DefaultResponseTimeout,
	LogLevel:        zapcore.DebugLevel.String(),
	LogFormat:       log.ConsoleFormat.String(),
}

func (f *Flags) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&f.ListenAddresses,
		"listeners",
		defaultFlags.ListenAddresses,
		"list of TCP or UNIX listening addresses separated by comma",
	)

	fs.DurationVar(
		&f.KeepAlivePeriod,
		"keep-alive",
		defaultFlags.KeepAlivePeriod,
		"keep alive period for the TCP connection",
	)

	fs.DurationVar(
		&f.ResponseTimeout,
		"timeout",
		defaultFlags.ResponseTimeout,
		"timeout for sending response to the client",
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

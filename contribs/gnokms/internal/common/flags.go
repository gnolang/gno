// Package common contains common flag definitions, authentication key file
// management, and utility functions used across gnokms commands.
package common

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/log"
	sserver "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
	"go.uber.org/zap/zapcore"
)

type AuthFlags struct {
	AuthKeysFile string
}

func defaultAuthKeysFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		var derr error
		// Unable to get the user's config directory, fallback to the current directory.
		if dir, derr = os.Getwd(); derr != nil {
			panic("Unable to get any of the user or current directories: " + errors.Join(
				err, derr,
			).Error())
		}
	}
	return filepath.Join(dir, "gnokms/auth_keys.json")
}

func (f *AuthFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&f.AuthKeysFile,
		"auth-keys-file",
		defaultAuthKeysFile(),
		"path to the file containing the authentication keys (both own private and client public keys)",
	)
}

type ServerFlags struct {
	AuthFlags

	Listener        string
	KeepAlivePeriod time.Duration
	ResponseTimeout time.Duration
	LogLevel        string
	LogFormat       string
}

var defaultServerFlags = ServerFlags{
	Listener:        "tcp://127.0.0.1:26659",
	KeepAlivePeriod: sserver.DefaultKeepAlivePeriod,
	ResponseTimeout: sserver.DefaultResponseTimeout,
	LogLevel:        zapcore.InfoLevel.String(),
	LogFormat:       log.ConsoleFormat.String(),
}

func (f *ServerFlags) RegisterFlags(fs *flag.FlagSet) {
	f.AuthFlags.RegisterFlags(fs)

	fs.StringVar(
		&f.Listener,
		"listener",
		defaultServerFlags.Listener,
		"TCP (tcp://<ip>:<port>) or UNIX (unix://<path>) listener",
	)

	fs.DurationVar(
		&f.KeepAlivePeriod,
		"keep-alive",
		defaultServerFlags.KeepAlivePeriod,
		"keep alive period for the TCP connection",
	)

	fs.DurationVar(
		&f.ResponseTimeout,
		"timeout",
		defaultServerFlags.ResponseTimeout,
		"timeout for sending response to the client",
	)

	fs.StringVar(
		&f.LogLevel,
		"log-level",
		defaultServerFlags.LogLevel,
		"log level (debug|info|warn|error)",
	)

	fs.StringVar(
		&f.LogFormat,
		"log-format",
		defaultServerFlags.LogFormat,
		"log format (json|console)",
	)
}

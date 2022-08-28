package client

import (
	"fmt"
	"os"
)

type BaseOptions struct {
	Home                  string `flag:"home" help:"home directory"`
	Remote                string `flag:"remote" help:"remote node URL (default 127.0.0.1:26657)"`
	Quiet                 bool   `flag:"quiet" help:"for parsing output"`
	InsecurePasswordStdin bool   `flag:"insecure-password-stdin" help:"WARNING! take password from stdin"`
}

var DefaultBaseOptions = BaseOptions{
	Home:                  homeDir(),
	Remote:                "127.0.0.1:26657",
	Quiet:                 false,
	InsecurePasswordStdin: false,
}

func homeDir() string {
	// if environment set, always use that.
	hd := os.Getenv("GNO_HOME")
	if hd != "" {
		return hd
	}

	// look for dir in home directory.
	hd, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/.gno", hd)
}

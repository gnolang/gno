package client

import (
	"fmt"
	"os"
)

type BaseOptions struct {
	Home   string `flag:"home" help:"home directory"`
	Remote string `flag:"remote" help:"remote node URL (default 127.0.0.1:26657)"`
	Quiet  bool   `flag:"quiet" help:"for parsing output"`
}

var DefaultBaseOptions = BaseOptions{
	Home:   homeDir(),
	Remote: "127.0.0.1:26657",
	Quiet:  false,
}

func homeDir() string {
	// if environment set, always use that.
	hd := os.Getenv("GNO_HOME")
	if hd != "" {
		return fmt.Sprintf("%s/.gno", hd)
	}
	// look for dir in local directory.
	// XXX
	// look for dir in home directory.
	hd, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/.gno", hd)
}

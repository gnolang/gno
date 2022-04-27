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
	hd, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/.gno", hd)
}

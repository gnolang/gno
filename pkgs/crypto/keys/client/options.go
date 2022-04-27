package client

import "os"

type BaseOptions struct {
	Home   string `flag:"home" help:"home directory"`
	Remote string `flag:"remote" help:"remote node URL (default 127.0.0.1:26657)"`
	Quiet  bool   `flag:"quiet" help:"for parsing output"`
}

var DefaultBaseOptions = BaseOptions{
	Home:   os.Getenv("GNO_HOME"),
	Remote: "127.0.0.1:26657",
	Quiet:  false,
}

package client

import (
	"fmt"
	"os"
)

type BaseOptions struct {
	Home                  string
	Remote                string
	Quiet                 bool
	InsecurePasswordStdin bool
}

var DefaultBaseOptions = BaseOptions{
	Home:                  HomeDir(),
	Remote:                "127.0.0.1:26657",
	Quiet:                 false,
	InsecurePasswordStdin: false,
}

func HomeDir() string {
	// if environment set, always use that.
	// if not, check $XDG_CONFIG_HOME
	// if not, fall back to home directory
	hd := os.Getenv("GNO_HOME")
	if hd != "" {
		return hd
	}
	hd = os.Getenv("XDG_CONFIG_HOME")
	if hd != "" {
		return fmt.Sprintf("%s/gno", hd)
	}
	hd, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/.gno", hd)
}

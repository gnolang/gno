package common

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

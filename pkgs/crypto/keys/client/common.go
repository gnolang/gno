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
	// if not, check whether can get os.UserHomeDir()
	// if not, fall back to home directory
	var err error
	hd := os.Getenv("GNO_HOME")
	if hd != "" {
		return hd
	}
	hd, err = os.UserConfigDir()
	if err == nil {
		return filepath.Join(hd, "gno")
	}
	hd, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/.gno", hd)
}

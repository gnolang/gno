package client

import (
	"fmt"
	"os"
	"path/filepath"
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
	dir := os.Getenv("GNO_HOME")
	if dir != "" {
		return dir
	}
	dir, err = os.UserConfigDir()
	if err == nil {
		return filepath.Join(dir, "gno")
	}
	dir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/.gno", dir)
}

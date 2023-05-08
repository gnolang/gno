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
	// if environment variable is set, always use that.
	// otherwise, use config dir (varies depending on OS) + "gno"
	var err error
	dir := os.Getenv("GNO_HOME")
	if dir != "" {
		return dir
	}
	dir, err = os.UserConfigDir()
	if err != nil {
		panic(fmt.Errorf("couldn't get user config dir: %w", err))
	}
	gnoHome := filepath.Join(dir, "gno")
	// XXX: added april 2023 as a transitory measure - remove after test4
	fixOldDefaultGnoHome(gnoHome)
	return gnoHome
}

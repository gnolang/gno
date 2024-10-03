package gnoenv

import (
	"fmt"
	"os"
	"path/filepath"
)

func HomeDir() string {
	// if environment variable is set, always use that.
	// otherwise, use config dir (varies depending on OS) + "gno"

	dir := os.Getenv("GNOHOME")
	if dir != "" {
		return dir
	}

	// XXX: `GNO_HOME` is deprecated and should be replaced by `GNOHOME`
	// keeping for compatibility support
	dir = os.Getenv("GNO_HOME")
	if dir != "" {
		return dir
	}

	var err error
	dir, err = os.UserConfigDir()
	if err != nil {
		panic(fmt.Errorf("couldn't get user config dir: %w", err))
	}
	gnoHome := filepath.Join(dir, "gno")

	// XXX: added april 2023 as a transitory measure - remove after test4
	fixOldDefaultGnoHome(gnoHome)
	return gnoHome
}

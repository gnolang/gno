package client

import (
	"fmt"
	"log"
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

func fixOldDefaultGnoHome(newDir string) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	oldDir := filepath.Join(dir, ".gno")
	s, err := os.Stat(oldDir)
	if err != nil || !s.IsDir() {
		return
	}
	if err = os.Rename(oldDir, newDir); err != nil {
		if os.IsExist(err) {
			log.Printf("WARNING: attempted moving old default GNO_HOME (%q) to new (%q) but failed because directory exists.", oldDir, newDir)
			log.Printf("You may need to move files from the old directory manually, or set the env var GNO_HOME to %q to retain the old directory.", oldDir)
		} else {
			log.Printf("WARNING: attempted moving old default GNO_HOME (%q) to new (%q) but failed with error: %v", oldDir, newDir, err)
		}
	}
}

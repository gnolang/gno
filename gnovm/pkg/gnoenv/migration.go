package gnoenv

import (
	"log"
	"os"
	"path/filepath"
)

// XXX: added april 2023 as a transitory measure - remove after test4
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

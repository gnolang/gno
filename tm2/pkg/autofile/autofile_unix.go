//go:build !js && !wasm
// +build !js,!wasm

package autofile

import (
	"os"
	"os/signal"
	"syscall"
)

func (af *AutoFile) setupCloseHandler() error {
	// Close file on SIGHUP.
	af.hupc = make(chan os.Signal, 1)
	signal.Notify(af.hupc, syscall.SIGHUP)
	go func() {
		for range af.hupc {
			af.closeFile()
		}
	}()
	return nil
}

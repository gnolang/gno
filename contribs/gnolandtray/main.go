package main

import (
	"log"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"

	"github.com/gnolang/gno/gno.land/pkg/integration"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Gno.land") // TODO: use a small icon instead of a title.
	systray.SetTooltip("Local Gno.land Node Manager")

	mHelp := systray.AddMenuItem("Help", "Help")
	mQuit := systray.AddMenuItem("Quit", "Quit")

	_ = integration.TestingInMemoryNode
	//node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	//println(node, remoteAddr)

	go func() {
		for {
			select {
			case <-mHelp.ClickedCh:
				open.Run("https://github.com/gnolang/gno/tree/master/contribs/gnolandtray")
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// clean up here
	log.Println("Exited.")
}

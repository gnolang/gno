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
	systray.SetTitle("Gnodev") // TODO: use a small icon instead of a title.
	systray.SetTooltip("Local Gno.land Node Manager")

	mStartGnodev := systray.AddMenuItem("Start Gnodev...", "")
	mOpenGnolandRPC := systray.AddMenuItem("Open Gnoland RPC in browser", "")
	mOpenGnoweb := systray.AddMenuItem("Open Gnoweb in browser", "")
	mOpenFolder := systray.AddMenuItem("Open Gnodev Folder", "")
	mHelp := systray.AddMenuItem("Help", "Help")
	mQuit := systray.AddMenuItem("Quit", "Quit")

	// show git sha version
	// show port
	// show metrics (memory, txs, height, etc)
	// 'open at login'
	// check for update, recommend rebuilding
	// "reset realms' state"
	// "save archive/dump"
	// "[ ] debug mode"

	_ = integration.TestingInMemoryNode
	//node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	//println(node, remoteAddr)

	go func() {
		for {
			select {
			case <-mStartGnodev.ClickedCh:
				mStartGnodev.SetTitle("Starting...")
				log.Println("NOT IMPLEMENTED")
			case <-mOpenGnolandRPC.ClickedCh:
				// open.Run("http://127.0.0.1:XXX/")
				log.Println("NOT IMPLEMENTED")
			case <-mOpenGnoweb.ClickedCh:
				// open.Run("http://127.0.0.1:XXX/")
				log.Println("NOT IMPLEMENTED")
			case <-mOpenFolder.ClickedCh:
				// open.Open("./...")
				log.Println("NOT IMPLEMENTED")
			case <-mHelp.ClickedCh:
				open.Run("https://github.com/gnolang/gno/tree/master/contribs/gnolandtray")
			case <-mQuit.ClickedCh:
				mStartGnodev.SetTitle("Stopping...")
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

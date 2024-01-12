package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	gnodev "github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	tmlog "github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

func main() {
	systray.Run(onReady, onExit)
}

const (
	startGnodevStr = "Start Gnodev..."
	stopGnodevStr  = "Stop Gnodev..."
	openGnolandStr = "Open Gnoland RPC in browser"
	openGnowebStr  = "Open Gnoweb in browser"
	openGnodevStr  = "Open Gnodev Folder"
	helpStr        = "Help"
	quitStr        = "Quit"
)

func onReady() {
	var (
		gnowebListener net.Listener
		node           *gnodev.Node
		nodeCancel     context.CancelCauseFunc = func(error) {} // noop
	)

	systray.SetTitle("Gnodev 👋") // TODO: use a small icon instead of a title.
	systray.SetTooltip("Local Gno.land Node Manager")

	// TODO: when ready -> green dot
	// TODO: when error -> red dot

	mStartGnodev := systray.AddMenuItem(startGnodevStr, "")
	mStopGnodev := systray.AddMenuItem(stopGnodevStr, "")
	mStopGnodev.Disable()
	mOpenGnolandRPC := systray.AddMenuItem(openGnolandStr, "")
	mOpenGnolandRPC.Disable()
	mOpenGnoweb := systray.AddMenuItem(openGnowebStr, "")
	mOpenGnoweb.Disable()
	mOpenFolder := systray.AddMenuItem(openGnodevStr, "")
	mOpenGnoweb.Disable()
	systray.AddSeparator()
	// mSettings := systray.AddMenuItem("Settings", "Settings")
	// mSettings.AddSubMenuItemCheckbox("Open at login", "TODO", false)
	// mSettings.AddSubMenuItemCheckbox("Debug/Verbose", "TODO", false)
	mHelp := systray.AddMenuItem(helpStr, "")
	mQuit := systray.AddMenuItem(quitStr, "")

	// show git sha version
	// show port
	// show metrics (memory, txs, height, etc)
	// check for update, recommend rebuilding
	// "reset realms' state"
	// "save archive/dump"

	_ = integration.TestingInMemoryNode
	// node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	// println(node, remoteAddr)

	go func() {
		for {
			select {
			case <-mStartGnodev.ClickedCh:
				systray.SetTitle("Gnodev 🪫")
				mStartGnodev.SetTitle("Starting...")
				mStartGnodev.Disable()
				ctx, cancel := context.WithCancelCause(context.Background())
				nodeCancel = cancel
				defer cancel(nil)
				gnoroot := gnoenv.RootDir()
				examplesDir := filepath.Join(gnoroot, "examples")
				// pkgpaths, err := parseArgsPackages(args); if err
				// pkgpaths = append(pkgpaths, examplesDir)
				pkgpaths := []string{examplesDir}
				osm.TrapSignal(func() {
					cancel(nil)
				})
				nodeOut := os.Stdout
				logger := tmlog.NewTMLogger(nodeOut)
				logger.SetLevel(tmlog.LevelError)
				var err error
				node, err = gnodev.NewDevNode(ctx, logger, pkgpaths)
				if err != nil {
					panic(err)
				}

				log.Printf("Listener: %s\n", node.GetRemoteAddress())
				log.Printf("Default Address: %s\n", gnodev.DefaultCreator.String())
				log.Printf("Chain ID: %s\n", node.Config().ChainID())

				gnowebListen := ":8000" // TODO: make custom?

				// TODO: auto-reload

				// gnoweb
				gnowebListener, err = net.Listen("tcp", gnowebListen)
				if err != nil {
					panic(fmt.Errorf("unable to listen to %q: %w", gnowebListen, err))
				}
				serveGnoWebServer := func(l net.Listener, dnode *gnodev.Node) error {
					var server http.Server

					webConfig := gnoweb.NewDefaultConfig()
					webConfig.RemoteAddr = dnode.GetRemoteAddress()

					loggerweb := tmlog.NewTMLogger(os.Stdout)
					loggerweb.SetLevel(tmlog.LevelDebug)

					app := gnoweb.MakeApp(loggerweb, webConfig)

					server.ReadHeaderTimeout = 60 * time.Second
					server.Handler = app.Router

					if err := server.Serve(l); err != nil {
						return fmt.Errorf("unable to serve GnoWeb: %w", err)
					}

					return nil
				}
				go func() {
					cancel(serveGnoWebServer(gnowebListener, node))
				}()
				log.Printf("Listener: http://%s\n", gnowebListener.Addr().String())

				go func() {
					started := time.Now()
					ticker := time.NewTicker(1 * time.Second)
					defer ticker.Stop()

					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							since := formatDuration(time.Since(started))
							mStartGnodev.SetTitle("Running... " + since)
						}
					}
				}()
				mStartGnodev.SetTitle("Running...")
				mStopGnodev.Enable()
				systray.SetTitle("Gnodev 🟢")
			case <-mStopGnodev.ClickedCh:
				mStopGnodev.Disable()
				mStopGnodev.SetTitle("Stopping...")
				nodeCancel(nil)
				node.Close()
				mStopGnodev.SetTitle("Stopped")

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

func formatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	seconds := totalSeconds % 60
	minutes := (totalSeconds / 60) % 60
	hours := totalSeconds / 3600

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

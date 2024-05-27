package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gnolang/gno/gno.me/apps"
	"github.com/gnolang/gno/gno.me/event"
	"github.com/gnolang/gno/gno.me/event/subscription"
	"github.com/gnolang/gno/gno.me/gno"
	gnohttp "github.com/gnolang/gno/gno.me/http"
	"github.com/gnolang/gno/gno.me/state"
	"github.com/gnolang/gno/gno.me/ws"
)

type Instance struct {
	vm                  gno.VM
	wsManager           *ws.Manager
	httpServer          *http.Server
	eventCh             chan *state.Event
	eventProcessingDone chan struct{}
	managerDone         chan struct{}
	wsServer            *event.Server
	wsServerDone        chan struct{}
	shutdown            bool
}

func Start(httpPort, wsPort string) *Instance {
	fmt.Println("Initializing VM...")
	vm, isFirstStartup := gno.NewVM()
	eventCh := make(chan *state.Event, 128)
	eventProcessingDone := make(chan struct{})
	managerDone := make(chan struct{})
	wsServerDone := make(chan struct{})

	var wsManager *ws.Manager
	if wsPort != "" {
		fmt.Println("Initializing incoming events manager...")
		wsManager = ws.NewManager(eventCh, managerDone)

		fmt.Println("Initializing event listeners and subscription channels...")
		memPackages := vm.QueryMemPackages(context.Background())
		if memPackages != nil {
			for memPackage := range memPackages {

				// Initialize the event listeners for syncable packages installed from remote sources.
				if memPackage.Address != "" && memPackage.Syncable {
					gno.RemoteApps.Add(memPackage.Name)

					// TODO: handle this error.
					if err := wsManager.SubscribeToPackageEvents(memPackage.Address, memPackage.Name); err != nil {
						fmt.Printf("error listening on package %s at address %s: %v\n", memPackage.Name, memPackage.Address, err)
					}
				}

				// If the package was created by this app, create the channel for subscribers.
				if memPackage.Syncable {
					subscription.AddChannel(memPackage.Name)
				}
			}
		}
	}

	fmt.Println("Initializing HTTP server...")
	httpServer := gnohttp.NewServerWithRemoteSupport(vm, wsManager, httpPort, wsPort)

	if isFirstStartup {
		fmt.Println("Creating installer apps...")
		if err := apps.CreatePort(vm); err != nil {
			panic(err.Error())
		}
		apps.CreateInstaller(vm)
		apps.CreateRemoteInstaller(vm)
	}

	// Overwrite the existing port number on each startup.
	if _, _, err := vm.Call(context.Background(), "port", false, "Set", httpPort); err != nil {
		panic("error setting port: " + err.Error())
	}

	fmt.Println("Starting HTTP server...")
	go httpServer.ListenAndServe()

	var eventServer *event.Server
	if wsPort != "" {
		fmt.Println("Starting event processing...")
		eventProcessor := event.NewProcessor(vm, eventCh, eventProcessingDone)
		go eventProcessor.Process()

		fmt.Println("Starting WS server for remote app requests and event broadcasting...")
		eventServer = event.NewServer(eventCreator{vm: vm}, vm, wsServerDone)
		eventServer.Start(wsPort)
	} else {
		close(eventProcessingDone)
		close(managerDone)
		close(wsServerDone)
	}

	instance := &Instance{
		vm:                  vm,
		wsManager:           wsManager,
		httpServer:          httpServer,
		eventCh:             eventCh,
		eventProcessingDone: eventProcessingDone,
		managerDone:         managerDone,
		wsServer:            eventServer,
		wsServerDone:        wsServerDone,
	}

	fmt.Println("READY :)")
	return instance
}

func (i *Instance) Stop() {
	if i.shutdown {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	i.httpServer.Shutdown(ctx)

	if i.wsServer != nil {
		i.wsServer.Stop()
		i.wsManager.Stop()
	}

	<-i.wsServerDone
	<-i.managerDone
	close(i.eventCh)
	<-i.eventProcessingDone

	i.shutdown = true
}

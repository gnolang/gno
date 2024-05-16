package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gnolang/gno/gno.me/apps"
	"github.com/gnolang/gno/gno.me/gno"
	gnohttp "github.com/gnolang/gno/gno.me/http"
	"github.com/gnolang/gno/gno.me/ws"
)

type Instance struct {
	vm                  gno.VM
	wsManager           *ws.Manager
	httpServer          *http.Server
	eventCh             chan *gno.Event
	stopEventProcessing chan struct{}
	eventProcessingDone chan struct{}
	managerDone         chan struct{}
	shutdown            bool
}

func Start(httpPort string) *Instance {
	fmt.Println("Initializing VM...")
	vm, isFirstStartup := gno.NewVM()
	eventCh := make(chan *gno.Event, 128)
	stopEventProcessing := make(chan struct{})
	eventProcessingDone := make(chan struct{})
	managerDone := make(chan struct{})

	fmt.Println("Initializing incoming events manager...")
	wsManager := ws.NewManager(eventCh, managerDone)

	fmt.Println("Initializing HTTP server...")
	httpServer := gnohttp.NewServerWithRemoteSupport(vm, wsManager, httpPort)

	// Get the list of remote apps from the VM and initiate all remote event listeners.
	fmt.Println("Initializing remote app event listeners...")
	memPackages := vm.QueryRemoteMemPackages(context.Background())
	if memPackages != nil {
		for memPackage := range memPackages {
			fmt.Println(memPackage.Name)
			if memPackage.Address == "" {
				continue
			}

			// TODO: handle this error.
			if err := wsManager.ListenOnPackage(memPackage.Address, memPackage.Path); err != nil {
				fmt.Printf("error listening on package %s at address %s: %v\n", memPackage.Path, memPackage.Address, err)
			}
		}
	}

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

	go func() {
		fmt.Println("Starting HTTP server...")
		httpServer.ListenAndServe()
	}()

	instance := &Instance{
		vm:                  vm,
		wsManager:           wsManager,
		httpServer:          httpServer,
		eventCh:             eventCh,
		stopEventProcessing: stopEventProcessing,
		eventProcessingDone: eventProcessingDone,
		managerDone:         managerDone,
	}

	fmt.Println("Starting event processing...")
	go instance.processEvents()

	return instance
}

func (i *Instance) Stop() {
	if i.shutdown {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	i.httpServer.Shutdown(ctx)
	i.wsManager.Stop()

	i.stopEventProcessing <- struct{}{}
	<-i.managerDone
	close(i.eventCh)
	<-i.eventProcessingDone
	close(i.stopEventProcessing)

	i.shutdown = true
}

func (i *Instance) processEvents() {
LOOP:
	for {
		select {
		case <-i.stopEventProcessing:
			break LOOP
		case event := <-i.eventCh:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			// TODO: handle these errors better. If this fails do to the event being out of order, then it should
			// initiate requests to get the necessary previous events.
			if err := i.vm.ApplyEvent(ctx, event); err != nil {
				fmt.Println("error applying event:", err)
			}

			cancel()
		}
	}

	// Finish processing events.
	for event := range i.eventCh {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		if err := i.vm.ApplyEvent(ctx, event); err != nil {
			fmt.Println("error applying event:", err)
		}

		cancel()
	}

	close(i.eventProcessingDone)
}

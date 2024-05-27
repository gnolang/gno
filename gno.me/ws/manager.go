package ws

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gnolang/gno/gno.me/state"
)

var (
	ErrShuttingDown            = errors.New("manager is shutting down")
	ErrConnectionAlreadyExists = errors.New("connection already exists for this package")
)

type Manager struct {
	sync.RWMutex
	done           chan struct{}
	conns          map[string]*connection
	connectionDone chan *connection
	eventCh        chan *state.Event
	stopCh         chan struct{}
	shuttingDown   bool
}

func NewManager(eventCh chan *state.Event, done chan struct{}) *Manager {
	manager := Manager{
		done:           done,
		conns:          make(map[string]*connection),
		connectionDone: make(chan *connection, 10),
		eventCh:        eventCh,
		stopCh:         make(chan struct{}),
	}

	go manager.Manage()

	return &manager
}

func (m *Manager) Manage() {
	for {
		select {
		case conn := <-m.connectionDone:
			m.Lock()
			if !m.shuttingDown {
				newConnection(conn.address, conn.appName, m.eventCh, m.connectionDone)
			}
			m.Unlock()
		case <-m.stopCh:
			m.Lock()
			m.shuttingDown = true
			m.Unlock()
			close(m.done)
			return
		}
	}
}

func (m *Manager) SubscribeToPackageEvents(address, appName string) error {
	m.Lock()
	defer m.Unlock()

	if m.shuttingDown {
		return ErrShuttingDown
	}

	if _, ok := m.conns[appName]; ok {
		return ErrConnectionAlreadyExists
	}

	conn, err := newConnection(address, appName, m.eventCh, m.connectionDone)
	if err != nil {
		return err
	}

	m.conns[appName] = conn
	return nil
}

func (m *Manager) Stop() {
	close(m.stopCh)
}

func (m *Manager) SubmitEvent(event *state.Event) error {
	m.RLock()
	defer m.RUnlock()

	conn, ok := m.conns[event.AppName]
	if !ok {
		fmt.Println("could not find connection for app", event.AppName)
		return nil
	}

	return conn.submitRemote(event)
}

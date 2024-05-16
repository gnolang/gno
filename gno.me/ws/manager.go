package ws

import (
	"errors"
	"sync"

	"github.com/gnolang/gno/gno.me/gno"
)

var (
	ErrShuttingDown            = errors.New("manager is shutting down")
	ErrConnectionAlreadyExists = errors.New("connection already exists for this package")
)

type Manager struct {
	sync.Mutex
	done           chan struct{}
	conns          map[string]*connection
	connectionDone chan *connection
	eventCh        chan *gno.Event
	stopCh         chan struct{}
	shuttingDown   bool
}

func NewManager(eventCh chan *gno.Event, done chan struct{}) *Manager {
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
				newConnection(conn.address, conn.pkgPath, m.eventCh, m.connectionDone)
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

func (m *Manager) ListenOnPackage(address, pkgPath string) error {
	m.Lock()
	defer m.Unlock()

	if m.shuttingDown {
		return ErrShuttingDown
	}

	if _, ok := m.conns[pkgPath]; ok {
		return ErrConnectionAlreadyExists
	}

	if err := newConnection(address, pkgPath, m.eventCh, m.connectionDone); err != nil {
		return err
	}

	m.conns[pkgPath] = &connection{address: address, pkgPath: pkgPath}
	return nil
}

func (m *Manager) Stop() {
	close(m.stopCh)
}

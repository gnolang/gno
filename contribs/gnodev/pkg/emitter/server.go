package emitter

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	cachepath "github.com/gnolang/gno/contribs/gnodev/pkg/cachepath"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gorilla/websocket"
)

type Emitter interface {
	Emit(evt events.Event)
}

type Server struct {
	logger    *slog.Logger
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]struct{}
	muClients sync.RWMutex
}

func NewServer(logger *slog.Logger) *Server {
	return &Server{
		logger:  logger,
		clients: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // XXX: adjust this
			},
		},
	}
}

func (s *Server) LockEmit() { s.muClients.Lock() }

func (s *Server) UnlockEmit() { s.muClients.Unlock() }

// ws handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("unable to upgrade connection", "remote", r.RemoteAddr, "error", err)
		return
	}
	defer conn.Close()

	s.muClients.Lock()
	s.clients[conn] = struct{}{}
	s.muClients.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.muClients.Lock()
			delete(s.clients, conn)
			s.muClients.Unlock()
			break
		}
	}
}

func (s *Server) Emit(evt events.Event) {
	go s.emit(evt)
}

type EventJSON struct {
	Type events.Type `json:"type"`
	Data any         `json:"data"`
}

func (s *Server) emit(evt events.Event) {
	s.muClients.Lock()
	defer s.muClients.Unlock()

	s.logEvent(evt)

	jsonEvt := EventJSON{evt.Type(), evt}
	for conn := range s.clients {
		err := conn.WriteJSON(jsonEvt)
		if err != nil {
			s.logger.Error("write json event", "error", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

func (s *Server) conns() []*websocket.Conn {
	s.muClients.RLock()
	conns := make([]*websocket.Conn, 0, len(s.clients))
	for conn := range s.clients {
		conns = append(conns, conn)
	}
	s.muClients.RUnlock()

	return conns
}

func (s *Server) logEvent(evt events.Event) {
	var logEvt string
	if rawEvt, err := json.Marshal(evt); err == nil {
		logEvt = string(rawEvt)

		if evt.Type() == events.EvtTxResult {
			type TxMsg struct {
				Package struct {
					Path string `json:"path"`
					Name string `json:"name"`
				} `json:"package"`
			}
			type Tx struct {
				Msg []TxMsg `json:"msg"`
			}
			type TxResultJSON struct {
				Tx Tx `json:"tx"`
			}

			var txResult TxResultJSON
			if err := json.Unmarshal(rawEvt, &txResult); err == nil && len(txResult.Tx.Msg) > 0 {
				packagePath := txResult.Tx.Msg[0].Package.Path
				packageName := txResult.Tx.Msg[0].Package.Name
				s.logger.Info("User addPkg ", "name", packageName,
					"path", packagePath)
				cachepath.Set(packagePath)
			} else {
				s.logger.Warn("Failed to parse package path or no messages found", "error", err)
			}
		}
	}

	s.logger.Info("sending event to clients",
		"clients", len(s.clients),
		"type", evt.Type(),
		"event", logEvt)
}

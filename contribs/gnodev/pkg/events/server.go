package events

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slog"
)

type Emitter interface {
	Emit(evt *Event)
}

type Server struct {
	logger    *slog.Logger
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]struct{}
	muClients sync.RWMutex
}

func NewEmitterServer(logger *slog.Logger) *Server {
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

func (s *Server) Emit(evt *Event) {
	go s.emit(evt)
}

func (s *Server) emit(evt *Event) {
	s.muClients.RLock()
	defer s.muClients.RUnlock()

	s.logger.Info("sending event to clients", "clients", len(s.clients), "event", evt.Type, "data", evt.Data)
	if len(s.clients) == 0 {
		return
	}

	for conn := range s.clients {
		err := conn.WriteJSON(evt)
		if err != nil {
			s.logger.Error("write json event", "error", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

var tmplFuncs = template.FuncMap{
	"jsEventsArray": func(events []EventType) string {
		var b strings.Builder
		b.WriteString("[")
		for i, v := range events {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%q", v))
		}
		b.WriteString("]")
		return b.String()
	},
}

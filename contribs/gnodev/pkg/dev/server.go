package dev

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gorilla/websocket"
)

type Emitter interface {
	Emit(evt *events.Event)
}

type EmitterServer struct {
	logger    log.Logger
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]struct{}
	muClients sync.Mutex
}

func NewEmitterServer(logger log.Logger) *EmitterServer {
	return &EmitterServer{
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
func (s *EmitterServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("unable to upgrade connection", "error", err)
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

func (s *EmitterServer) Emit(evt *events.Event) {
	if len(s.clients) == 0 {
		return
	}

	s.muClients.Lock()
	defer s.muClients.Unlock()

	s.logger.Info("sending json", "clients", len(s.clients), "event", evt.Type, "data", evt.Data)
	for conn := range s.clients {
		err := conn.WriteJSON(evt)
		if err != nil {
			s.logger.Error("write json", "error", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

var tmplFuncs = template.FuncMap{
	"jsEventsArray": func(events []events.EventType) string {
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

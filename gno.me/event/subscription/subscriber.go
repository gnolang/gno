package subscription

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Subscriber struct {
	sync.Mutex
	id          uuid.UUID
	conn        *websocket.Conn
	broadcastTo bool
}

func NewSubscriber(conn *websocket.Conn) *Subscriber {
	return &Subscriber{
		id:   uuid.New(),
		conn: conn,
	}
}

func (s *Subscriber) ID() uuid.UUID {
	return s.id
}

func (s *Subscriber) Send(msg []byte) error {
	return s.conn.WriteMessage(websocket.TextMessage, msg)
}

func (s *Subscriber) SetBroadcastTo(broadcastTo bool) {
	s.Lock()
	defer s.Unlock()

	s.broadcastTo = broadcastTo
}

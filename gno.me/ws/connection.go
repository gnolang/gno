package ws

import (
	"fmt"

	"github.com/gnolang/gno/gno.me/gno"
	"github.com/gorilla/websocket"
)

const (
	MsgTypeListenOnPackage int = iota
)

type connection struct {
	address string
	pkgPath string
}

func newConnection(
	address string,
	pkgPath string,
	eventCh chan *gno.Event,
	done chan *connection,
) error {
	conn, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		return fmt.Errorf("could not connect to address %s for package %s: %w", address, pkgPath, err)
	}

	if err := conn.WriteMessage(MsgTypeListenOnPackage, []byte(pkgPath)); err != nil {
		conn.Close()
		return fmt.Errorf("could not send message to address %s for package %s: %w", address, pkgPath, err)
	}

	go func() {
		defer conn.Close()

		for {
			var event gno.Event
			if err := conn.ReadJSON(&event); err != nil {
				done <- &connection{address: address, pkgPath: pkgPath}
				return
			}

			eventCh <- &event
		}
	}()

	return nil
}

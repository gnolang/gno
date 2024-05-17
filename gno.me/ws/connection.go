package ws

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gnolang/gno/gno.me/event/message"
	"github.com/gnolang/gno/gno.me/state"
	"github.com/gorilla/websocket"
)

type connection struct {
	sync.Mutex
	address string
	appName string
	conn    *websocket.Conn
}

func newConnection(
	address string,
	appName string,
	eventCh chan *state.Event,
	done chan *connection,
) (*connection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(address+"/events", nil)
	if err != nil {
		return nil, fmt.Errorf("could not connect to address %s for package %s: %w", address, appName, err)
	}

	fmt.Println("listening for events from remote app", appName, "at address", address+"/events")

	var initMsg message.Init
	initMsg.AppName = appName
	msg, err := initMsg.Marshal()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not marshal app name for package %s: %w", appName, err)
	}

	fmt.Println("sending it message")
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not send init message to address %s for package %s: %w", address, appName, err)
	}

	var subMsg message.Subscribe
	subMsg.AppName = appName
	msg, err = subMsg.Marshal()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not marshal subscribe message for package %s: %w", appName, err)
	}

	// TODO: do not subscribe first. First catch up to the latest event.
	fmt.Println("sending subscribe message")
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not send message to address %s for package %s: %w", address, appName, err)
	}

	// Catch up to the latest event.
	// if err := syncToLatestEvent(conn, address, appName, eventCh); err != nil {
	// 	conn.Close()
	// 	return nil, fmt.Errorf("could not sync to latest event for address %s for package %s: %w", address, appName, err)
	// }

	go func() {
		defer conn.Close()

		for {
			var genericMsg message.Generic
			if err := conn.ReadJSON(&genericMsg); err != nil {
				done <- &connection{address: address, appName: appName}
				return
			}

			var sendMsg message.Send
			if err := json.Unmarshal(genericMsg.Payload, &sendMsg); err != nil {
				fmt.Println("error unmarshalling send message:", err)
				continue
			}

			// TODO: check if the event sequence matches what we'd expect. Enter catch-up mode if it doesn't.

			fmt.Println("received event; sending to event channel")
			fmt.Println("event:", *sendMsg.Event)
			eventCh <- sendMsg.Event
		}
	}()

	return &connection{
		address: address,
		appName: appName,
		conn:    conn,
	}, nil
}

// func syncToLatestEvent(conn *websocket.Conn, address, appName string, eventCh chan *state.Event) error {
// 	return nil
// }

func (c *connection) submitRemote(event *state.Event) error {
	c.Lock()
	defer c.Unlock()

	submitMsg := message.Submit{Event: event}
	msg, err := submitMsg.Marshal()
	if err != nil {
		return err
	}

	fmt.Println("writing submit message for remote app", event.AppName)
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

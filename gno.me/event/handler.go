package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gnolang/gno/gno.me/event/message"
	"github.com/gnolang/gno/gno.me/event/subscription"
	"github.com/gnolang/gno/gno.me/state"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{}
	server   Server
)

type NameReq struct {
	AppName string `json:"app_name"`
}

func handleEvents(resp http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("unable to upgrade connection"))
		return
	}
	defer conn.Close()

	fmt.Println("event server connection initiated")

	// Read the first message used to initiate the connection and discard it.
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("error reading init message:", err)
		return
	}

	genericMsg, err := message.UnmarshalGeneric(msg)
	if err != nil {
		fmt.Println("error unmarshalling generic message:", err)
		return
	}

	if genericMsg.Type != message.TypeInit {
		fmt.Println("invalid message type:", genericMsg.Type)
		return
	}

	var initMsg message.Init
	if err = json.Unmarshal(genericMsg.Payload, &initMsg); err != nil {
		fmt.Println("error unmarshalling name request:", err)
		return
	}

	channel := subscription.GetChannel(initMsg.AppName)
	if channel == nil {
		fmt.Println("channel not found", initMsg.AppName)
		return
	}

	fmt.Println("creating new subscriber")
	subscriber := subscription.NewSubscriber(conn)
	channel.AddSubscriber(subscriber)

	for {
		err = conn.ReadJSON(&genericMsg)
		if err != nil {
			readErr := err
			errMsg, err := message.ReadErr{Message: err.Error()}.Marshal()
			if err != nil {
				fmt.Println("error marshalling read error message:", err)
				break
			}

			if err = conn.WriteMessage(websocket.TextMessage, errMsg); err != nil {
				fmt.Println("error writing read error message:", err)
				break
			}

			fmt.Println("error reading message:", readErr)
			continue
		}

		var (
			events  []*state.Event
			channel *subscription.Channel
		)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		fmt.Println("handling message of type", genericMsg.Type)
		switch genericMsg.Type {
		case message.TypeSubmit:
			fmt.Println("handling submit event")
			events, channel, err = handleSubmitEvent(ctx, genericMsg.Payload)
			if err != nil {
				fmt.Println("error handling submit event:", err)
				continue // TODO: handle error
			}
		case message.TypeSubscribe:
			fmt.Println("handling subscribe event")
			subscriber.SetBroadcastTo(true)
			continue
		case message.TypeRequest:
			//events, err = handleMsgRequestEvents(ctx, msg)
		default:
			continue
		}

		for _, event := range events {
			// msg, err = json.Marshal(event)
			// if err != nil {
			// 	fmt.Println("error marshalling event for broadcast", err)
			// 	break
			// }

			fmt.Println("broadcasting event")
			failed, err := channel.Broadcast(event)
			if err != nil {
				fmt.Println("error broadcasting event:", err)
				break
			}

			if len(failed) > 0 {
				channel.RemoveSubscribers(failed)
			}
		}
	}
}

func handleSubmitEvent(ctx context.Context, msg []byte) ([]*state.Event, *subscription.Channel, error) {
	var msgSubmit message.Submit
	fmt.Println(string(msg))
	if err := json.Unmarshal(msg, &msgSubmit); err != nil {
		return nil, nil, err
	}

	event := msgSubmit.Event
	channel := subscription.GetChannel(event.AppName)
	if channel == nil {
		return nil, nil, errors.New("app not found: " + event.AppName)
	}

	events, err := server.eventCreator.CreateEvents(ctx, event.AppName, event.Func, event.Args...)
	return events, channel, err
}

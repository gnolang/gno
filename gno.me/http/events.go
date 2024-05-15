package http

import (
	"context"
	"encoding/json"
	gohttp "net/http"
	"time"

	"github.com/gnolang/gno/gno.me/gno"
	"github.com/gorilla/websocket"
)

const (
	MsgTypeInit int = iota
	MsgTypeCall
	MsgTypeReadErr
	MsgTypeRequestEvents
	MsgTypeSendEvents
	MsgTypeEventEncodingErr
)

var upgrader = websocket.Upgrader{}

func handleEvents(resp gohttp.ResponseWriter, req *gohttp.Request) {
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		resp.WriteHeader(gohttp.StatusInternalServerError)
		resp.Write([]byte("unable to upgrade connection"))
		return
	}
	defer conn.Close()

	// Read the first message used to initiate the connection and discard it.
	if _, _, err = conn.ReadMessage(); err != nil {
		resp.WriteHeader(gohttp.StatusInternalServerError)
		resp.Write([]byte("unable to initiate connection"))
		return
	}

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if err = conn.WriteMessage(MsgTypeReadErr, []byte("error reading message")); err != nil {
				resp.WriteHeader(gohttp.StatusInternalServerError)
				resp.Write([]byte("error reading and writing to websocket; closing conection"))
				break
			}
			continue
		}

		var events []gno.Event
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		switch msgType {
		case MsgTypeCall:
			events, err = handleMsgCall(ctx, msg)
		case MsgTypeRequestEvents:
			//events, err = handleMsgRequestEvents(ctx, msg)
		default:
			continue
		}

		msg, err = json.Marshal(events)
		if err != nil {
			if err = conn.WriteMessage(MsgTypeEventEncodingErr, []byte{}); err != nil {
				resp.WriteHeader(gohttp.StatusInternalServerError)
				resp.Write([]byte("error writing to websocket; closing conection"))
			}
		}

		if err = conn.WriteMessage(MsgTypeSendEvents, msg); err != nil {
			resp.WriteHeader(gohttp.StatusInternalServerError)
			resp.Write([]byte("error writing to websocket; closing conection"))
		}
	}
}

func handleMsgCall(ctx context.Context, msg []byte) ([]gno.Event, error) {
	var msgCall gno.MsgCall
	if err := json.Unmarshal(msg, &msgCall); err != nil {
		return nil, err
	}

	_, events, err := vm.Call(ctx, msgCall.AppName, false, msgCall.Func, msgCall.Args...)
	return events, err
}

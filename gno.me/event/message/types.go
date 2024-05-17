package message

import (
	"encoding/json"

	"github.com/gnolang/gno/gno.me/state"
)

const (
	// Client Message Types
	TypeInit int = iota
	TypeSubscribe
	TypeLatestSequence
	TypeRequest
	TypeSubmit

	// Server Message Types
	TypeReadErr
	TypeSend
	TypeEncodingErr
)

type Generic struct {
	Type    int    `json:"type"`
	Payload []byte `json:"payload"`
}

func UnmarshalGeneric(b []byte) (Generic, error) {
	var g Generic
	if err := json.Unmarshal(b, &g); err != nil {
		return Generic{}, err
	}

	return g, nil
}

type Init struct {
	AppName string `json:"app_name"`
}

func (i Init) Marshal() ([]byte, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	return json.Marshal(Generic{Type: TypeInit, Payload: b})
}

type Subscribe struct {
	AppName string `json:"app_name"`
}

func (s Subscribe) Marshal() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return json.Marshal(Generic{Type: TypeSubscribe, Payload: b})
}

type Submit struct {
	Event *state.Event `json:"event"`
}

func (s Submit) Marshal() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return json.Marshal(Generic{Type: TypeSubmit, Payload: b})
}

type ReadErr struct {
	Message string `json:"message"`
}

func (r ReadErr) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return json.Marshal(Generic{Type: TypeReadErr, Payload: b})
}

type Send struct {
	Event *state.Event `json:"event"`
}

func (s Send) Marshal() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return json.Marshal(Generic{Type: TypeSend, Payload: b})
}

type EncodingErr struct {
	Message string `json:"message"`
}

func (e EncodingErr) Marshal() ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return json.Marshal(Generic{Type: TypeEncodingErr, Payload: b})
}

package types_test

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/aib/ibc/types"
)

func TestMsgSendPacketValidateBasic(t *testing.T) {
	var msg *types.MsgSendPacket
	var payload types.Payload
	testCases := []struct {
		name     string
		malleate func()
		expErr   string
	}{
		{
			name:     "success",
			malleate: func() {},
		},
		{
			name: "success, multiple payloads",
			malleate: func() {
				msg.Payloads = append(msg.Payloads, payload)
			},
		},
		{
			name: "failure: empty source client",
			malleate: func() {
				msg.SourceClient = ""
			},
			expErr: "validate client identifier: identifier cannot be blank",
		},
		{
			name: "failure: invalid timestamp",
			malleate: func() {
				msg.TimeoutTimestamp = 0
			},
			expErr: "timeout must not be 0",
		},
		{
			name: "failure: invalid length for payload",
			malleate: func() {
				msg.Payloads = []types.Payload{}
			},
			expErr: "payload length must be greater than 0",
		},
		{
			name: "failure: invalid packetdata",
			malleate: func() {
				msg.Payloads = []types.Payload{{}}
			},
			expErr: "invalid payload #0: invalid source port",
		},
		{
			name: "failure: invalid payload",
			malleate: func() {
				msg.Payloads[0].DestinationPort = ""
			},
			expErr: "invalid payload #0: invalid destination port",
		},
		{
			name: "failure: invalid multiple payload",
			malleate: func() {
				payload.DestinationPort = ""
				msg.Payloads = append(msg.Payloads, payload)
			},
			expErr: "invalid payload #1: invalid destination port",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload = types.Payload{
				SourcePort:      "sourcePort",
				DestinationPort: "destinationPort",
				Version:         "ics20-1",
				Encoding:        "encoding",
				Value:           []byte("packetData"),
			}
			msg = types.NewMsgSendPacket(
				"sourceClient", uint64(time.Now().Add(time.Hour).Unix()), payload,
			)
			tc.malleate()

			err := msg.ValidateBasic()

			if tc.expErr == "" && err != nil {
				t.Errorf("expected no error, got %s", err)
				return
			}
			if tc.expErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expErr) {
					t.Errorf("expected error %s, got %s", tc.expErr, err)
				}
			}
		})
	}
}

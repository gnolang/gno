package types_test

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/urequire"
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

func TestPacketValidateBasic(t *testing.T) {
	var packet types.Packet
	var payload types.Payload
	testCases := []struct {
		name     string
		malleate func()
		expErr   string
	}{
		{
			"success",
			func() {},
			"",
		},
		{
			"success, single payload just below MaxPayloadsSize",
			func() {
				packet.Payloads[0].Value = make([]byte, types.MaximumPayloadsSize-1)
			},
			"",
		},
		{
			"success, multiple payloads",
			func() {
				packet.Payloads = append(packet.Payloads, payload)
			},
			"",
		},
		{
			"failure: invalid single payloads size",
			func() {
				// bytes that are larger than MaxPayloadsSize
				packet.Payloads[0].Value = make([]byte, types.MaximumPayloadsSize+1)
			},
			"packet data bytes cannot exceed",
		},
		{
			"failure: invalid total payloads size",
			func() {
				payload.Value = make([]byte, types.MaximumPayloadsSize-1)
				packet.Payloads = append(packet.Payloads, payload)
			},
			"packet data bytes cannot exceed",
		},
		{
			"failure: payloads is nil",
			func() {
				packet.Payloads = nil
			},
			"payload length must be greater than 0",
		},
		{
			"failure: empty payload",
			func() {
				packet.Payloads = []types.Payload{}
			},
			"payload length must be greater than 0",
		},
		{
			"failure: invalid payload source port ID",
			func() {
				packet.Payloads[0].SourcePort = ""
			},
			"invalid Payload #0: invalid source port",
		},
		{
			"failure: invalid payload dest port ID",
			func() {
				packet.Payloads[0].DestinationPort = ""
			},
			"invalid Payload #0: invalid destination port",
		},
		{
			"failure: invalid source ID",
			func() {
				packet.SourceClient = ""
			},
			"invalid source ID",
		},
		{
			"failure: invalid dest ID",
			func() {
				packet.DestinationClient = ""
			},
			"invalid destination ID",
		},
		{
			"failure: invalid sequence",
			func() {
				packet.Sequence = 0
			},
			"packet sequence cannot be 0",
		},
		{
			"failure: invalid timestamp",
			func() {
				packet.TimeoutTimestamp = 0
			},
			"packet timeout timestamp cannot be 0",
		},
		{
			"failure: empty version",
			func() {
				packet.Payloads[0].Version = ""
			},
			"payload version cannot be empty",
		},
		{
			"failure: empty encoding",
			func() {
				packet.Payloads[0].Encoding = ""
			},
			"payload encoding cannot be empty",
		},
		{
			"failure: empty value",
			func() {
				packet.Payloads[0].Value = []byte{}
			},
			"payload value cannot be empty",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload = types.Payload{
				SourcePort:      "sourcePort",
				DestinationPort: "destinationPort",
				Version:         "ics20-v2",
				Encoding:        "application/json",
				Value:           []byte("{}"),
			}
			packet = types.NewPacket(1, "channel-1", "channel-2", uint64(time.Now().Unix()), payload)
			tc.malleate()

			err := packet.ValidateBasic()

			if tc.expErr == "" {
				urequire.NoError(t, err)
				return
			}
			urequire.ErrorContains(t, err, tc.expErr)
		})
	}
}

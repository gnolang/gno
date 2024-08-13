package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/gnolang/gnostats/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	errInvalidRegisterRequest = errors.New("invalid register request")
	errInvalidInfoAddress     = errors.New("invalid info address")
	errInvalidInfoGnoVersion  = errors.New("invalid info gno version")
	errInvalidInfoOSVersion   = errors.New("invalid info OS version")

	errHubClosed         = errors.New("hub closed")
	errUnregisteredAgent = errors.New("unregistered agent")
)

// Hub is the
type Hub struct {
	proto.UnimplementedHubServer

	agents sync.Map // address -> static info
	subs   subs     // the active data subs
}

// NewHub creates a new hub instance
func NewHub() *Hub {
	return &Hub{
		subs: make(subs),
	}
}

// Register registers the node instance with the stats hub
func (h *Hub) Register(_ context.Context, info *proto.StaticInfo) (*emptypb.Empty, error) {
	// TODO add handshake
	// Sanity check the request
	if err := verifyStaticInfo(info); err != nil {
		return nil, err
	}

	// Register the agent,
	// overwriting the previous entry, if any
	h.agents.Store(info.Address, info)

	return &emptypb.Empty{}, nil
}

// GetDataStream returns a stream of fresh data from the stats hub
func (h *Hub) GetDataStream(_ *emptypb.Empty, stream proto.Hub_GetDataStreamServer) error {
	// Create a subscription
	id, ch := h.subs.subscribe()
	defer h.subs.unsubscribe(id)

	// Grab the stream context
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		case data, more := <-ch:
			if !more {
				return errHubClosed
			}

			// Forward the data point
			if err := stream.Send(data); err != nil {
				return fmt.Errorf("unable to stream data point, %w", err)
			}
		}
	}
}

// PushData continuously pushes the node data to the stats hub
func (h *Hub) PushData(stream proto.Hub_PushDataServer) error {
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Grab the data point from the agent
			data, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					return err
				}

				return nil
			}

			// Fetch the info
			info, registered := h.agents.Load(data.Address)
			if !registered {
				return errUnregisteredAgent
			}

			// Prepare the data point
			parsed := &proto.DataPoint{
				DynamicInfo: data,
				StaticInfo:  info.(*proto.StaticInfo),
			}

			// Notify the listeners, if any
			h.subs.notify(parsed)
		}
	}
}

// verifyStaticInfo verifies the node's static info
func verifyStaticInfo(info *proto.StaticInfo) error {
	// Check if the request was initialized
	if info == nil {
		return errInvalidRegisterRequest
	}

	// Check if the address was set
	if info.Address == "" {
		return errInvalidInfoAddress
	}

	// Check if the gno version was set
	if info.GnoVersion == "" {
		return errInvalidInfoGnoVersion
	}

	// Check if the OS version was set
	if info.OsVersion == "" {
		return errInvalidInfoOSVersion
	}

	// Location info can be uninitialized
	return nil
}

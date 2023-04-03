package p2p

import (
	"fmt"
	"net"
)

// FilterTimeoutError indicates that a filter operation timed out.
type FilterTimeoutError struct{}

func (e FilterTimeoutError) Error() string {
	return "filter timed out"
}

// RejectedError indicates that a Peer was rejected carrying additional
// information as to the reason.
type RejectedError struct {
	addr              NetAddress
	conn              net.Conn
	err               error
	id                ID
	isAuthFailure     bool
	isDuplicate       bool
	isFiltered        bool
	isIncompatible    bool
	isNodeInfoInvalid bool
	isSelf            bool
}

// Addr returns the NetAddress for the rejected Peer.
func (e RejectedError) Addr() NetAddress {
	return e.addr
}

func (e RejectedError) Error() string {
	if e.isAuthFailure {
		return fmt.Sprintf("auth failure: %s", e.err)
	}

	if e.isDuplicate {
		if e.conn != nil {
			return fmt.Sprintf(
				"duplicate CONN<%s>",
				e.conn.RemoteAddr().String(),
			)
		}
		if !e.id.IsZero() {
			return fmt.Sprintf("duplicate ID<%v>", e.id)
		}
	}

	if e.isFiltered {
		if e.conn != nil {
			return fmt.Sprintf(
				"filtered CONN<%s>: %s",
				e.conn.RemoteAddr().String(),
				e.err,
			)
		}

		if !e.id.IsZero() {
			return fmt.Sprintf("filtered ID<%v>: %s", e.id, e.err)
		}
	}

	if e.isIncompatible {
		return fmt.Sprintf("incompatible: %s", e.err)
	}

	if e.isNodeInfoInvalid {
		return fmt.Sprintf("invalid NodeInfo: %s", e.err)
	}

	if e.isSelf {
		return fmt.Sprintf("self ID<%v>", e.id)
	}

	return fmt.Sprintf("%s", e.err)
}

// IsAuthFailure when Peer authentication was unsuccessful.
func (e RejectedError) IsAuthFailure() bool { return e.isAuthFailure }

// IsDuplicate when Peer ID or IP are present already.
func (e RejectedError) IsDuplicate() bool { return e.isDuplicate }

// IsFiltered when Peer ID or IP was filtered.
func (e RejectedError) IsFiltered() bool { return e.isFiltered }

// IsIncompatible when Peer NodeInfo is not compatible with our own.
func (e RejectedError) IsIncompatible() bool { return e.isIncompatible }

// IsNodeInfoInvalid when the sent NodeInfo is not valid.
func (e RejectedError) IsNodeInfoInvalid() bool { return e.isNodeInfoInvalid }

// IsSelf when Peer is our own node.
func (e RejectedError) IsSelf() bool { return e.isSelf }

// SwitchDuplicatePeerIDError to be raised when a peer is connecting with a known
// ID.
type SwitchDuplicatePeerIDError struct {
	ID ID
}

func (e SwitchDuplicatePeerIDError) Error() string {
	return fmt.Sprintf("duplicate peer ID %v", e.ID)
}

// SwitchDuplicatePeerIPError to be raised when a peer is connecting with a known
// IP.
type SwitchDuplicatePeerIPError struct {
	IP net.IP
}

func (e SwitchDuplicatePeerIPError) Error() string {
	return fmt.Sprintf("duplicate peer IP %v", e.IP.String())
}

// SwitchConnectToSelfError to be raised when trying to connect to itself.
type SwitchConnectToSelfError struct {
	Addr *NetAddress
}

func (e SwitchConnectToSelfError) Error() string {
	return fmt.Sprintf("connect to self: %v", e.Addr)
}

type SwitchAuthenticationFailureError struct {
	Dialed *NetAddress
	Got    ID
}

func (e SwitchAuthenticationFailureError) Error() string {
	return fmt.Sprintf(
		"failed to authenticate peer. Dialed %v, but got peer with ID %s",
		e.Dialed,
		e.Got,
	)
}

// TransportClosedError is raised when the Transport has been closed.
type TransportClosedError struct{}

func (e TransportClosedError) Error() string {
	return "transport has been closed"
}

// -------------------------------------------------------------------

type NetAddressNoIDError struct {
	Addr string
}

func (e NetAddressNoIDError) Error() string {
	return fmt.Sprintf("address (%s) does not contain ID", e.Addr)
}

type NetAddressInvalidError struct {
	Addr string
	Err  error
}

func (e NetAddressInvalidError) Error() string {
	return fmt.Sprintf("invalid address (%s): %v", e.Addr, e.Err)
}

type NetAddressLookupError struct {
	Addr string
	Err  error
}

func (e NetAddressLookupError) Error() string {
	return fmt.Sprintf("error looking up host (%s): %v", e.Addr, e.Err)
}

// CurrentlyDialingOrExistingAddressError indicates that we're currently
// dialing this address or it belongs to an existing peer.
type CurrentlyDialingOrExistingAddressError struct {
	Addr string
}

func (e CurrentlyDialingOrExistingAddressError) Error() string {
	return fmt.Sprintf("connection with %s has been established or dialed", e.Addr)
}

// Package discovery contains the p2p peer discovery service (Reactor).
// The purpose of the peer discovery service is to gather peer lists from known peers,
// and attempt to fill out open peer connection slots in order to build out a fuller mesh.
//
// The implementation of the peer discovery protocol is relatively simple.
// In essence, it pings a random peer at a specific interval (3s), for a list of their known peers (max 30).
// After receiving the list, and verifying it, the node attempts to establish outbound connections to the
// given peers.
package discovery

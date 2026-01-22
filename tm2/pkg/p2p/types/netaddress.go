// Modified for Tendermint
// Originally Copyright (c) 2013-2014 Conformal Systems LLC.
// https://github.com/conformal/btcd/blob/master/LICENSE

package types

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

const (
	nilNetAddress = "<nil-NetAddress>"
	badNetAddress = "<bad-NetAddress>"
)

var (
	ErrInvalidTCPAddress = errors.New("invalid TCP address")
	ErrUnsetIPAddress    = errors.New("unset IP address")
	ErrInvalidIP         = errors.New("invalid IP address")
	ErrUnspecifiedIP     = errors.New("unspecified IP address")
	ErrInvalidNetAddress = errors.New("invalid net address")
	ErrEmptyHost         = errors.New("empty host address")
)

// NetAddress defines information about a peer on the network
// including its ID, IP address, and port
type NetAddress struct {
	ID       ID     `json:"id"`                 // unique peer identifier (public key address)
	IP       net.IP `json:"ip"`                 // the IP part of the dial address
	Hostname string `json:"hostname,omitempty"` // original hostname, if any
	Port     uint16 `json:"port"`               // the port part of the dial address
}

// NetAddressString returns id@addr. It strips the leading
// protocol from protocolHostPort if it exists.
func NetAddressString(id ID, protocolHostPort string) string {
	return fmt.Sprintf(
		"%s@%s",
		id,
		removeProtocolIfDefined(protocolHostPort),
	)
}

// NewNetAddress returns a new NetAddress using the provided TCP
// address
func NewNetAddress(id ID, addr net.Addr) (*NetAddress, error) {
	// Make sure the address is valid
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return nil, ErrInvalidTCPAddress
	}

	// Validate the ID
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("unable to verify ID, %w", err)
	}

	na := NewNetAddressFromIPPort(
		tcpAddr.IP,
		uint16(tcpAddr.Port),
	)

	// Set the ID
	na.ID = id

	return na, nil
}

// NewNetAddressFromString returns a new NetAddress using the provided address in
// the form of "ID@IP:Port".
// Also resolves the host if host is not an IP.
func NewNetAddressFromString(idaddr string) (*NetAddress, error) {
	var (
		prunedAddr = removeProtocolIfDefined(idaddr)
		spl        = strings.Split(prunedAddr, "@")
	)

	if len(spl) != 2 {
		return nil, ErrInvalidNetAddress
	}

	var (
		id   = crypto.ID(spl[0])
		addr = spl[1]
	)

	// Validate the ID
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("unable to verify address ID, %w", err)
	}

	// Extract the host and port
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("unable to split host and port, %w", err)
	}

	if host == "" {
		return nil, ErrEmptyHost
	}

	hostname := host

	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, fmt.Errorf("unable to look up IP, %w", err)
		}

		ip = ips[0]
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("unable to parse port %s, %w", portStr, err)
	}

	na := NewNetAddressFromIPPort(ip, uint16(port))
	na.Hostname = hostname
	na.ID = id

	return na, nil
}

// NewNetAddressFromStrings returns an array of NetAddress'es build using
// the provided strings.
func NewNetAddressFromStrings(idaddrs []string) ([]*NetAddress, []error) {
	var (
		netAddrs = make([]*NetAddress, 0, len(idaddrs))
		errs     = make([]error, 0, len(idaddrs))
	)

	for _, addr := range idaddrs {
		netAddr, err := NewNetAddressFromString(addr)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		netAddrs = append(netAddrs, netAddr)
	}

	return netAddrs, errs
}

// NewNetAddressFromIPPort returns a new NetAddress using the provided IP
// and port number.
func NewNetAddressFromIPPort(ip net.IP, port uint16) *NetAddress {
	hostname := ""
	if ip != nil {
		hostname = ip.String() // preserve original IP string as hostname
	}

	return &NetAddress{
		IP:       ip,
		Hostname: hostname,
		Port:     port,
	}
}

// Equals reports whether na and other are the same addresses,
// including their ID, IP, and Port.
func (na *NetAddress) Equals(other NetAddress) bool {
	return na.String() == other.String()
}

// Same returns true is na has the same non-empty ID or DialString as other.
func (na *NetAddress) Same(other NetAddress) bool {
	var (
		dialsSame = na.DialString() == other.DialString()
		IDsSame   = na.ID != "" && na.ID == other.ID
	)

	return dialsSame || IDsSame
}

// String representation: <ID>@<IP>:<PORT>
func (na *NetAddress) String() string {
	if na == nil {
		return nilNetAddress
	}

	str, err := na.MarshalAmino()
	if err != nil {
		return badNetAddress
	}

	return str
}

// MarshalAmino stringifies a NetAddress.
// Needed because (a) IP doesn't encode, and (b) the intend of this type is to
// serialize to a string anyways.
func (na NetAddress) MarshalAmino() (string, error) {
	addrStr := na.DialString()

	if na.ID != "" {
		return NetAddressString(na.ID, addrStr), nil
	}

	return addrStr, nil
}

func (na *NetAddress) UnmarshalAmino(raw string) (err error) {
	netAddress, err := NewNetAddressFromString(raw)
	if err != nil {
		return err
	}

	*na = *netAddress

	return nil
}

func (na *NetAddress) DialString() string {
	if na == nil {
		return nilNetAddress
	}

	return net.JoinHostPort(
		na.IP.String(),
		strconv.FormatUint(uint64(na.Port), 10),
	)
}

// ResolveIP resolves the hostname for the address (if any) and updates the IP
// field with the latest lookup result.
func (na *NetAddress) ResolveIP(ctx context.Context) error {
	if na == nil || na.Hostname == "" {
		return nil
	}

	if ip := net.ParseIP(na.Hostname); ip != nil {
		na.IP = ip

		return nil
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, na.Hostname)
	if err != nil {
		return fmt.Errorf("unable to resolve host %s, %w", na.Hostname, err)
	}

	if len(addrs) == 0 {
		return fmt.Errorf("unable to resolve host %s, no addresses found", na.Hostname)
	}

	na.IP = addrs[0].IP

	return nil
}

// DialContext dials the given NetAddress with a context
func (na *NetAddress) DialContext(ctx context.Context) (net.Conn, error) {
	var d net.Dialer

	conn, err := d.DialContext(ctx, "tcp", na.DialString())
	if err != nil {
		return nil, fmt.Errorf("unable to dial address, %w", err)
	}

	return conn, nil
}

// Routable returns true if the address is routable.
func (na *NetAddress) Routable() bool {
	if err := na.Validate(); err != nil {
		return false
	}

	// TODO(oga) bitcoind doesn't include RFC3849 here, but should we?
	return !(na.RFC1918() ||
		na.RFC3927() ||
		na.RFC4862() ||
		na.RFC4193() ||
		na.RFC4843() ||
		na.Local())
}

// Validate validates the NetAddress params
func (na *NetAddress) Validate() error {
	// Validate the ID
	if err := na.ID.Validate(); err != nil {
		return fmt.Errorf("unable to validate ID, %w", err)
	}

	// Make sure the IP is set
	if na.IP == nil {
		return ErrUnsetIPAddress
	}

	// Make sure the IP is valid
	ipLen := len(na.IP)
	if ipLen != 4 && ipLen != 16 {
		return ErrInvalidIP
	}

	// Check if the IP is unspecified
	if na.IP.IsUnspecified() {
		return ErrUnspecifiedIP
	}

	// Check if the IP conforms to standards, or is a broadcast
	if na.RFC3849() || na.IP.Equal(net.IPv4bcast) {
		return ErrInvalidIP
	}

	return nil
}

// Local returns true if it is a local address.
func (na *NetAddress) Local() bool {
	return na.IP.IsLoopback() || zero4.Contains(na.IP)
}

// RFC1918: IPv4 Private networks (10.0.0.0/8, 192.168.0.0/16, 172.16.0.0/12)
// RFC3849: IPv6 Documentation address  (2001:0DB8::/32)
// RFC3927: IPv4 Autoconfig (169.254.0.0/16)
// RFC3964: IPv6 6to4 (2002::/16)
// RFC4193: IPv6 unique local (FC00::/7)
// RFC4380: IPv6 Teredo tunneling (2001::/32)
// RFC4843: IPv6 ORCHID: (2001:10::/28)
// RFC4862: IPv6 Autoconfig (FE80::/64)
// RFC6052: IPv6 well known prefix (64:FF9B::/96)
// RFC6145: IPv6 IPv4 translated address ::FFFF:0:0:0/96
var rfc1918_10 = net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)}

var (
	rfc1918_192 = net.IPNet{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)}
	rfc1918_172 = net.IPNet{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)}
	rfc3849     = net.IPNet{IP: net.ParseIP("2001:0DB8::"), Mask: net.CIDRMask(32, 128)}
	rfc3927     = net.IPNet{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)}
	rfc3964     = net.IPNet{IP: net.ParseIP("2002::"), Mask: net.CIDRMask(16, 128)}
	rfc4193     = net.IPNet{IP: net.ParseIP("FC00::"), Mask: net.CIDRMask(7, 128)}
	rfc4380     = net.IPNet{IP: net.ParseIP("2001::"), Mask: net.CIDRMask(32, 128)}
	rfc4843     = net.IPNet{IP: net.ParseIP("2001:10::"), Mask: net.CIDRMask(28, 128)}
	rfc4862     = net.IPNet{IP: net.ParseIP("FE80::"), Mask: net.CIDRMask(64, 128)}
	rfc6052     = net.IPNet{IP: net.ParseIP("64:FF9B::"), Mask: net.CIDRMask(96, 128)}
	rfc6145     = net.IPNet{IP: net.ParseIP("::FFFF:0:0:0"), Mask: net.CIDRMask(96, 128)}
	zero4       = net.IPNet{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(8, 32)}
)

func (na *NetAddress) RFC1918() bool {
	return rfc1918_10.Contains(na.IP) ||
		rfc1918_192.Contains(na.IP) ||
		rfc1918_172.Contains(na.IP)
}
func (na *NetAddress) RFC3849() bool { return rfc3849.Contains(na.IP) }
func (na *NetAddress) RFC3927() bool { return rfc3927.Contains(na.IP) }
func (na *NetAddress) RFC3964() bool { return rfc3964.Contains(na.IP) }
func (na *NetAddress) RFC4193() bool { return rfc4193.Contains(na.IP) }
func (na *NetAddress) RFC4380() bool { return rfc4380.Contains(na.IP) }
func (na *NetAddress) RFC4843() bool { return rfc4843.Contains(na.IP) }
func (na *NetAddress) RFC4862() bool { return rfc4862.Contains(na.IP) }
func (na *NetAddress) RFC6052() bool { return rfc6052.Contains(na.IP) }
func (na *NetAddress) RFC6145() bool { return rfc6145.Contains(na.IP) }

// removeProtocolIfDefined removes the protocol part of the given address
func removeProtocolIfDefined(addr string) string {
	if !strings.Contains(addr, "://") {
		// No protocol part
		return addr
	}

	return strings.Split(addr, "://")[1]
}

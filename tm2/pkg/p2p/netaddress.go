// Modified for Tendermint
// Originally Copyright (c) 2013-2014 Conformal Systems LLC.
// https://github.com/conformal/btcd/blob/master/LICENSE

package p2p

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

type ID = crypto.ID

// NetAddress defines information about a peer on the network
// including its Address, IP address, and port.
// NOTE: NetAddress is not meant to be mutated due to memoization.
// @amino2: immutable XXX
type NetAddress struct {
	ID   ID     `json:"id"`   // authenticated identifier (TODO)
	IP   net.IP `json:"ip"`   // part of "addr"
	Port uint16 `json:"port"` // part of "addr"

	// TODO:
	// Name string `json:"name"` // optional DNS name

	// memoize .String()
	str string
}

// NetAddressString returns id@addr. It strips the leading
// protocol from protocolHostPort if it exists.
func NetAddressString(id ID, protocolHostPort string) string {
	addr := removeProtocolIfDefined(protocolHostPort)
	return fmt.Sprintf("%s@%s", id, addr)
}

// NewNetAddress returns a new NetAddress using the provided TCP
// address. When testing, other net.Addr (except TCP) will result in
// using 0.0.0.0:0. When normal run, other net.Addr (except TCP) will
// panic. Panics if ID is invalid.
// TODO: socks proxies?
func NewNetAddress(id ID, addr net.Addr) *NetAddress {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		if flag.Lookup("test.v") == nil { // normal run
			panic(fmt.Sprintf("Only TCPAddrs are supported. Got: %v", addr))
		} else { // in testing
			netAddr := NewNetAddressFromIPPort("", net.IP("0.0.0.0"), 0)
			netAddr.ID = id
			return netAddr
		}
	}

	if err := id.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid ID %v: %v (addr: %v)", id, err, addr))
	}

	ip := tcpAddr.IP
	port := uint16(tcpAddr.Port)
	na := NewNetAddressFromIPPort("", ip, port)
	na.ID = id
	return na
}

// NewNetAddressFromString returns a new NetAddress using the provided address in
// the form of "ID@IP:Port".
// Also resolves the host if host is not an IP.
// Errors are of type ErrNetAddressXxx where Xxx is in (NoID, Invalid, Lookup)
func NewNetAddressFromString(idaddr string) (*NetAddress, error) {
	idaddr = removeProtocolIfDefined(idaddr)
	spl := strings.Split(idaddr, "@")
	if len(spl) != 2 {
		return nil, NetAddressNoIDError{idaddr}
	}

	// get ID
	id := crypto.ID(spl[0])
	if err := id.Validate(); err != nil {
		return nil, NetAddressInvalidError{idaddr, err}
	}
	addr := spl[1]

	// get host and port
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, NetAddressInvalidError{addr, err}
	}
	if len(host) == 0 {
		return nil, NetAddressInvalidError{
			addr,
			errors.New("host is empty"),
		}
	}

	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, NetAddressLookupError{host, err}
		}
		ip = ips[0]
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, NetAddressInvalidError{portStr, err}
	}

	na := NewNetAddressFromIPPort("", ip, uint16(port))
	na.ID = id
	return na, nil
}

// NewNetAddressFromStrings returns an array of NetAddress'es build using
// the provided strings.
func NewNetAddressFromStrings(idaddrs []string) ([]*NetAddress, []error) {
	netAddrs := make([]*NetAddress, 0)
	errs := make([]error, 0)
	for _, addr := range idaddrs {
		netAddr, err := NewNetAddressFromString(addr)
		if err != nil {
			errs = append(errs, err)
		} else {
			netAddrs = append(netAddrs, netAddr)
		}
	}
	return netAddrs, errs
}

// NewNetAddressIPPort returns a new NetAddress using the provided IP
// and port number.
func NewNetAddressFromIPPort(id ID, ip net.IP, port uint16) *NetAddress {
	return &NetAddress{
		ID:   id,
		IP:   ip,
		Port: port,
	}
}

// Equals reports whether na and other are the same addresses,
// including their ID, IP, and Port.
func (na *NetAddress) Equals(other interface{}) bool {
	if o, ok := other.(*NetAddress); ok {
		return na.String() == o.String()
	}
	return false
}

// Same returns true is na has the same non-empty ID or DialString as other.
func (na *NetAddress) Same(other interface{}) bool {
	if o, ok := other.(*NetAddress); ok {
		if na.DialString() == o.DialString() {
			return true
		}
		if na.ID != "" && na.ID == o.ID {
			return true
		}
	}
	return false
}

// String representation: <ID>@<IP>:<PORT>
func (na *NetAddress) String() string {
	if na == nil {
		return "<nil-NetAddress>"
	}
	if na.str != "" {
		return na.str
	}
	str, err := na.MarshalAmino()
	if err != nil {
		return "<bad-NetAddress>"
	}
	return str
}

// Needed because (a) IP doesn't encode, and (b) the intend of this type is to
// serialize to a string anyways.
func (na NetAddress) MarshalAmino() (string, error) {
	if na.str == "" {
		addrStr := na.DialString()
		if na.ID != "" {
			addrStr = NetAddressString(na.ID, addrStr)
		}
		na.str = addrStr
	}
	return na.str, nil
}

func (na *NetAddress) UnmarshalAmino(str string) (err error) {
	na2, err := NewNetAddressFromString(str)
	if err != nil {
		return err
	}
	*na = *na2
	return nil
}

func (na *NetAddress) DialString() string {
	if na == nil {
		return "<nil-NetAddress>"
	}
	return net.JoinHostPort(
		na.IP.String(),
		strconv.FormatUint(uint64(na.Port), 10),
	)
}

// Dial calls net.Dial on the address.
func (na *NetAddress) Dial() (net.Conn, error) {
	conn, err := net.Dial("tcp", na.DialString())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// DialTimeout calls net.DialTimeout on the address.
func (na *NetAddress) DialTimeout(timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", na.DialString(), timeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Routable returns true if the address is routable.
func (na *NetAddress) Routable() bool {
	if err := na.Validate(); err != nil {
		return false
	}
	// TODO(oga) bitcoind doesn't include RFC3849 here, but should we?
	return !(na.RFC1918() || na.RFC3927() || na.RFC4862() ||
		na.RFC4193() || na.RFC4843() || na.Local())
}

func (na *NetAddress) ValidateLocal() error {
	if err := na.ID.Validate(); err != nil {
		return err
	}
	if na.IP == nil {
		return errors.New("no IP")
	}
	if len(na.IP) != 4 && len(na.IP) != 16 {
		return fmt.Errorf("invalid IP bytes: %v", len(na.IP))
	}
	if na.RFC3849() || na.IP.Equal(net.IPv4bcast) {
		return errors.New("invalid IP", na.IP.IsUnspecified())
	}
	return nil
}

func (na *NetAddress) Validate() error {
	if err := na.ID.Validate(); err != nil {
		return err
	}
	if na.IP == nil {
		return errors.New("no IP")
	}
	if len(na.IP) != 4 && len(na.IP) != 16 {
		return fmt.Errorf("invalid IP bytes: %v", len(na.IP))
	}
	if na.IP.IsUnspecified() || na.RFC3849() || na.IP.Equal(net.IPv4bcast) {
		return errors.New("invalid IP", na.IP.IsUnspecified())
	}
	return nil
}

// HasID returns true if the address has an ID.
// NOTE: It does not check whether the ID is valid or not.
func (na *NetAddress) HasID() bool {
	return !na.ID.IsZero()
}

// Local returns true if it is a local address.
func (na *NetAddress) Local() bool {
	return na.IP.IsLoopback() || zero4.Contains(na.IP)
}

// ReachabilityTo checks whenever o can be reached from na.
func (na *NetAddress) ReachabilityTo(o *NetAddress) int {
	const (
		Unreachable = 0
		Default     = iota
		Teredo
		Ipv6_weak
		Ipv4
		Ipv6_strong
	)
	switch {
	case !na.Routable():
		return Unreachable
	case na.RFC4380():
		switch {
		case !o.Routable():
			return Default
		case o.RFC4380():
			return Teredo
		case o.IP.To4() != nil:
			return Ipv4
		default: // ipv6
			return Ipv6_weak
		}
	case na.IP.To4() != nil:
		if o.Routable() && o.IP.To4() != nil {
			return Ipv4
		}
		return Default
	default: /* ipv6 */
		var tunnelled bool
		// Is our v6 is tunnelled?
		if o.RFC3964() || o.RFC6052() || o.RFC6145() {
			tunnelled = true
		}
		switch {
		case !o.Routable():
			return Default
		case o.RFC4380():
			return Teredo
		case o.IP.To4() != nil:
			return Ipv4
		case tunnelled:
			// only prioritise ipv6 if we aren't tunnelling it.
			return Ipv6_weak
		}
		return Ipv6_strong
	}
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

func removeProtocolIfDefined(addr string) string {
	if strings.Contains(addr, "://") {
		return strings.Split(addr, "://")[1]
	}
	return addr
}

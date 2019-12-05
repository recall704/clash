// +build amd64 arm64 mips64

package tun

import (
	"fmt"
	"net"
	"time"

	"github.com/Dreamacro/clash/common/pool"
	"github.com/Dreamacro/clash/dns"
	"github.com/Dreamacro/clash/log"
	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/adapters/gonet"
	"github.com/google/netstack/tcpip/network/ipv4"
	"github.com/google/netstack/tcpip/network/ipv6"
	"github.com/google/netstack/tcpip/stack"
	"github.com/google/netstack/tcpip/transport/udp"
	"github.com/google/netstack/waiter"
	D "github.com/miekg/dns"
)

const (
	defaultTimeout = 5
)

var (
	ipv4Zero = tcpip.Address(net.IPv4zero.To4())
	ipv6Zero = tcpip.Address(net.IPv6zero.To16())
)

// DNSServer is DNS Server listening on tun devcice
type DNSServer struct {
	*dns.Server
	resolver *dns.Resolver

	stack         *stack.Stack
	tcpListener   net.Listener
	udpEndpoint   *dnsEndpoint
	udpEndpointID *stack.TransportEndpointID
	tcpip.NICID
}

type dnsEndpoint struct {
	stack.TransportEndpoint
	stack        *stack.Stack
	uniqueID     uint64
	udpForwarder *udp.Forwarder

	server *dns.Server
}

type connResponseWriter struct {
	*gonet.Conn
}

func newDNSEndpoint(s *stack.Stack, server *dns.Server) *dnsEndpoint {
	ep := &dnsEndpoint{
		uniqueID: s.UniqueID(),
		server:   server,
	}
	ep.udpForwarder = udp.NewForwarder(s, func(request *udp.ForwarderRequest) {
		var wq waiter.Queue
		ep, err := request.CreateEndpoint(&wq)
		if err != nil {
			return
		}
		conn := gonet.NewConn(&wq, ep)
		go func() {
			buffer := pool.BufPool.Get().([]byte)
			defer pool.BufPool.Put(buffer[:cap(buffer)])
			defer conn.Close()
			w := &connResponseWriter{Conn: conn}
			var msg D.Msg
			for {
				conn.SetDeadline(time.Now().Add(defaultTimeout * time.Second))
				// TODO: handle request larger than MTU
				n, err := conn.Read(buffer[:])
				if err != nil {
					break
				}
				msg.Unpack(buffer[:n])
				go server.ServeDNS(w, &msg)
			}
		}()
	})
	return ep
}

func (e *dnsEndpoint) UniqueID() uint64 {
	return e.uniqueID
}

func (e *dnsEndpoint) HandlePacket(r *stack.Route, id stack.TransportEndpointID, pkt tcpip.PacketBuffer) {
	e.udpForwarder.HandlePacket(r, id, pkt)
}

func (e *dnsEndpoint) HandleControlPacket(id stack.TransportEndpointID, typ stack.ControlType, extra uint32, pkt tcpip.PacketBuffer) {
}

func (e *dnsEndpoint) Close() {
}

func (e *dnsEndpoint) Wait() {

}

func (w *connResponseWriter) WriteMsg(msg *D.Msg) error {
	b, err := msg.Pack()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (w *connResponseWriter) TsigStatus() error {
	return nil
}
func (w *connResponseWriter) TsigTimersOnly(bool) {
	// Unsupported
}
func (w *connResponseWriter) Hijack() {
	// Unsupported
}

// CreateDNSServer create a dns server on given netstack
func CreateDNSServer(s *stack.Stack, resolver *dns.Resolver, ip net.IP, port int, nicID tcpip.NICID) (*DNSServer, error) {

	var v4 bool
	var err error

	address := tcpip.FullAddress{NIC: nicID, Port: uint16(port)}
	if ip.To4() != nil {
		v4 = true
		address.Addr = tcpip.Address(ip.To4())
	} else {
		address.Addr = tcpip.Address(ip.To16())
		v4 = false
	}
	if address.Addr == ipv4Zero || address.Addr == ipv6Zero {
		address.Addr = ""
	}

	handler := dns.NewHandler(resolver)
	serverIn := &dns.Server{}
	serverIn.SetHandler(handler)

	// UDP DNS

	id := &stack.TransportEndpointID{
		LocalAddress:  address.Addr,
		LocalPort:     uint16(port),
		RemotePort:    0,
		RemoteAddress: "",
	}
	endpoint := newDNSEndpoint(s, serverIn)

	if tcpiperr := s.RegisterTransportEndpoint(1,
		[]tcpip.NetworkProtocolNumber{
			ipv4.ProtocolNumber,
			ipv6.ProtocolNumber,
		},
		udp.ProtocolNumber,
		*id,
		endpoint,
		true,
		nicID); err != nil {
		log.Errorln("Unable to start UDP DNS on tun:  %v", tcpiperr.String())
	}

	// TCP DNS
	var tcpListener net.Listener
	if v4 {
		tcpListener, err = gonet.NewListener(s, address, ipv4.ProtocolNumber)
	} else {
		tcpListener, err = gonet.NewListener(s, address, ipv6.ProtocolNumber)
	}
	if err != nil {
		return nil, fmt.Errorf("Can not listen on tun: %v", err)
	}

	server := &DNSServer{
		Server:        serverIn,
		resolver:      resolver,
		stack:         s,
		tcpListener:   tcpListener,
		udpEndpoint:   endpoint,
		udpEndpointID: id,
		NICID:         nicID,
	}
	server.SetHandler(handler)
	server.Server.Server = &D.Server{Listener: tcpListener, Handler: server}

	go func() {
		server.ActivateAndServe()
	}()

	return server, err
}

// Stop stop the DNS Server on tun
func (s *DNSServer) Stop() {
	// shutdown TCP DNS Server
	s.Server.Shutdown()
	// remove TCP endpoint from stack
	if s.Listener != nil {
		s.Listener.Close()
	}
	// remove udp endpoint from stack
	s.stack.UnregisterTransportEndpoint(1,
		[]tcpip.NetworkProtocolNumber{
			ipv4.ProtocolNumber,
			ipv6.ProtocolNumber,
		},
		udp.ProtocolNumber,
		*s.udpEndpointID,
		s.udpEndpoint,
		s.NICID)
}

// DNSListen return the listening address of DNS Server
func (t *tunAdapter) DNSListen() string {
	if t.dnsserver != nil {
		id := t.dnsserver.udpEndpointID
		return fmt.Sprintf("%s:%d", id.LocalAddress.String(), id.LocalPort)
	}
	return ""
}

// Stop stop the DNS Server on tun
func (t *tunAdapter) ReCreateDNSServer(resolver *dns.Resolver, addr string) error {
	if addr == "" && t.dnsserver == nil {
		return nil
	}

	if addr == t.DNSListen() && t.dnsserver != nil && t.dnsserver.resolver == resolver {
		return nil
	}

	if t.dnsserver != nil {
		t.dnsserver.Stop()
		t.dnsserver = nil
		log.Debugln("Tun DNS server stoped")
	}

	var err error
	_, port, err := net.SplitHostPort(addr)
	if port == "0" || port == "" || err != nil {
		return nil
	}

	if resolver == nil {
		return fmt.Errorf("Failed to create DNS server on tun: resolver not provided")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	server, err := CreateDNSServer(t.ipstack, resolver, udpAddr.IP, udpAddr.Port, 1)
	if err != nil {
		return err
	}
	t.dnsserver = server
	log.Infoln("Tun DNS server listening at: %s", addr)
	return nil
}

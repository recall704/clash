package adapters

import (
	"net"

	"github.com/Dreamacro/clash/component/socks5"
	C "github.com/Dreamacro/clash/constant"
)

// TunAdapter is a adapter for socks and redir connection
type TunAdapter struct {
	net.Conn
	metadata *C.Metadata
}

// Metadata return destination metadata
func (s *TunAdapter) Metadata() *C.Metadata {
	return s.metadata
}

// NewTun is TunAdapter generator
func NewTun(target socks5.Addr, conn net.Conn, source C.Type, netType C.NetWork) *TunAdapter {
	metadata := parseSocksAddr(target)
	metadata.NetWork = netType
	metadata.Type = source
	if ip, port, err := parseAddr(conn.RemoteAddr().String()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = port
	}

	return &TunAdapter{
		Conn:     conn,
		metadata: metadata,
	}
}

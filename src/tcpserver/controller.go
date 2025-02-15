package tcpserver

import "net"

type TcpController interface {
	TcpNetworkAccept() bool
	RemoteAddrCheck(remoteAddr *net.TCPAddr) bool
}

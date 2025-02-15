package tcpserver

type TcpController interface {
	TcpNetworkAccept() bool
	RemoteAddrCheck(remoteAddr string) bool
}

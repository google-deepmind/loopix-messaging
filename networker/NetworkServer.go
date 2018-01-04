package networker

import "net"

type NetworkServer interface {
	ListenForIncomingConnections()
	HandleConnection(conn net.Conn)
	Run()
}

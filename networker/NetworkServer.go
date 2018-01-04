/*
	Package networker implements the interfaces for the Network (TCP) Client and Server.
*/

package networker

import "net"

type NetworkServer interface {
	ListenForIncomingConnections()
	HandleConnection(conn net.Conn)
	Run()
}

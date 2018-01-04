/*
	Package networker implements the interfaces for the Network (TCP) Client and Server.
 */

package networker

type NetworkClient interface {
	Send(packet string, host string, port string)
}

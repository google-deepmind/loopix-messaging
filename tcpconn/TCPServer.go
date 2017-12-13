package tcpconn

import "anonymous-messaging/packet_format"

type TCPServer interface {
	ReceivedPacket(packet packet_format.Packet)
	ForwardPacket(packet packet_format.Packet)
	SendPacket(packet packet_format.Packet)
	Run()
}

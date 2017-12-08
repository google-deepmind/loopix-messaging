package tcpconn

type TCPServer interface {
	ReceivedPacket(packet string)
	SendPacket(packet string)
	Start()
}

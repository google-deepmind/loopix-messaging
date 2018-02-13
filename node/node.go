package node

type Node interface {
	ProcessPacket(packet string)
	SendLoopMessage()
}

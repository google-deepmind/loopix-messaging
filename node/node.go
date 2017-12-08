package anonymous_messaging

type Node interface {
	ProcessPacket(packet string)
	SendLoopMessage()
}
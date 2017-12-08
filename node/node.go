package anonymous_messaging

type Node interface {
	ProcessPacket(p string) string
	SendLoopMessage()
	LogInfo()
}


package anonymous_messaging

type NodeOperations interface {
	ProcessPacket(p string) string
	SendLoopMessage()
	LogInfo()
}


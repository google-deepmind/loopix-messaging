package anonymous_messaging

import "fmt"

type NodeOperations interface {
	ProcessPacket(p string) string
	SendLoopMessage()
	LogInfo()
}

type Mix struct {
	Id string
	Host string
	IP string
	PubKey int
	PrvKey int
}

type Provider struct {
	Mix
}

func (m Mix) ProcessPacket(packet string) string {
	return packet
}

func (m Mix) SendLoopMessage() {
	fmt.Println("Sending loop message")
}

func (m Mix) LogInfo(msg string) {
	fmt.Println(msg)
}

func (p Provider) StorePacket(packet string) {
	fmt.Println(packet)
}


package anonymous_messaging

import "fmt"

type Mix struct {
	Id string
	PubKey int
	PrvKey int
	Server *TCPSocketServer
}

func (m Mix) StartMix() {
	m.Server.Start()
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
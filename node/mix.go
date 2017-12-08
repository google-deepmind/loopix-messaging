package anonymous_messaging

import (
	"fmt"
)

type Mix struct {
	Id string
	PubKey int
	PrvKey int
}


func (m Mix) ProcessPacket(packet string) {
	fmt.Println("Processing packet: ", packet)
}

func (m Mix) SendLoopMessage() {
	fmt.Println("Sending loop message")
}

func (m Mix) LogInfo(msg string) {
	fmt.Println(msg)
}
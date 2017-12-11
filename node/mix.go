package anonymous_messaging

import (
	"fmt"
	packet "anonymous-messaging/packet_format"
)

type Mix struct {
	Id string
	PubKey int
	PrvKey int
}


func (m Mix) ProcessPacket(p string, c chan<- string){
	fmt.Println("> Processing packet")

	dePacket := packet.FromString(p)
	c <- dePacket.ToString()
}

func (m Mix) SendLoopMessage() {
	fmt.Println("> Sending loop message")
}

func NewMix(id string, pubKey, prvKey int ) Mix{
	return Mix{Id:id, PubKey:pubKey, PrvKey:prvKey}
}
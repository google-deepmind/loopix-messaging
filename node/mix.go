package anonymous_messaging

import (
	"fmt"
	packet_format "anonymous-messaging/packet_format"
)

type Mix struct {
	Id string
	PubKey int
	PrvKey int
}


func (m Mix) ProcessPacket(p packet_format.Packet, c chan<- packet_format.Packet){
	fmt.Println("> Processing packet")

	dePacket := p
	c <- dePacket
}

func (m Mix) SendLoopMessage() {
	fmt.Println("> Sending loop message")
}

func NewMix(id string, pubKey, prvKey int ) Mix{
	return Mix{Id:id, PubKey:pubKey, PrvKey:prvKey}
}
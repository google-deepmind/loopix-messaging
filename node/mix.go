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

func NewMix(id string, pubKey, prvKey int ) Mix{
	mix := Mix{}
	mix.Id = id
	mix.PubKey = pubKey
	mix.PrvKey = prvKey
	return mix
}
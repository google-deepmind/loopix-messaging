package node

import (
	"anonymous-messaging/packet_format"
	"fmt"
	"time"
)

type Mix struct {
	Id     string
	PubKey int
	PrvKey int
}

func (m *Mix) ProcessPacket(p packet_format.Packet, c chan<- packet_format.Packet) {
	fmt.Println("> Processing packet")

	dePacket := packet_format.Decode(p)

	delay := dePacket.Steps[m.Id].Delay

	timeoutCh := make(chan packet_format.Packet, 1)

	go func(p packet_format.Packet, delay float64) {
		time.Sleep(time.Second * time.Duration(delay))
		timeoutCh <- p
	}(dePacket, delay)

	c <- <-timeoutCh
}

func (m *Mix) SendLoopMessage() {
	fmt.Println("> Sending loop message")
	// TO DO
}

func NewMix(id string, pubKey, prvKey int) *Mix {
	return &Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
}

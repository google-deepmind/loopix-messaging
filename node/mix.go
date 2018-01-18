/*
	Package node implements the core functions for a mix node, which allow to process the received cryptographic packets.
 */
package node

import (
	"fmt"

	sphinx "anonymous-messaging/new_packet_format"
	"anonymous-messaging/publics"
	"time"
)

type Mix struct {
	Id     string
	PubKey publics.PublicKey
	PrvKey publics.PrivateKey
}

func (m *Mix) ProcessPacket(packet []byte, c chan<- []byte, chop chan <- sphinx.Hop){

	nextHop, commands, newPacket, err := sphinx.ProcessSphinxPacket(packet, m.PrvKey)

	if err != nil {
		panic(err)
	}

	delay := commands.Delay
	fmt.Println(delay)
	timeoutCh := make(chan []byte, 1)

	go func(p []byte, delay float64) {
		time.Sleep(time.Second * time.Duration(delay))
		timeoutCh <- p
	}(newPacket.Bytes(), delay)

	c <- <-timeoutCh
	chop <- nextHop
}

func (m *Mix) SendLoopMessage() {
	fmt.Println("> Sending loop message")
}

func NewMix(id string, pubKey publics.PublicKey, prvKey publics.PrivateKey) *Mix {
	return &Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
}

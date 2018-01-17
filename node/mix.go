/*
	Package node implements the core functions for a mix node, which allow to process the received cryptographic packets.
 */
package node

import (
	"fmt"
	"time"

	sphinx "anonymous-messaging/new_packet_format"
	"anonymous-messaging/publics"
)

type Mix struct {
	Id     string
	PubKey publics.PublicKey
	PrvKey publics.PrivateKey
}

func (m *Mix) ProcessPacket(p sphinx.SphinxPacket, c chan<- sphinx.SphinxPacket, chop chan <- sphinx.Hop) {
	fmt.Println("> Processing packet")

	nextHop, commands, newPacket, err := sphinx.ProcessSphianxPacket(p, m.PrvKey)

	if err != nil {
		panic(err) // probably this should be a returned error and panic or logging of error should be done on a higher level
	}

	delay := commands.Delay

	timeoutCh := make(chan sphinx.SphinxPacket, 1)

	go func(p sphinx.SphinxPacket, delay float64) {
		time.Sleep(time.Second * time.Duration(delay))
		timeoutCh <- p
	}(newPacket, delay)

	c <- <-timeoutCh
	chop <- nextHop
}

func (m *Mix) SendLoopMessage() {
	fmt.Println("> Sending loop message")
	// TO DO
}

func NewMix(id string, pubKey publics.PublicKey, prvKey publics.PrivateKey) *Mix {
	return &Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
}

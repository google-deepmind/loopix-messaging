/*
	Package node implements the core functions for a mix node, which allow to process the received cryptographic packets.
 */
package node

import (
	"fmt"

	sphinx "anonymous-messaging/sphinx"
	"time"
)

type Mix struct {
	Id     string
	PubKey []byte //publics.PublicKey
	PrvKey []byte //publics.PrivateKey
}

func (m *Mix) ProcessPacket(packet []byte, c chan<- []byte, cAdr chan <- string, cFlag chan <- string){

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
	cAdr <- nextHop.Address
	cFlag <- commands.Flag
}

func (m *Mix) SendLoopMessage() {
	fmt.Println("> Sending loop message")
}

func NewMix(id string, pubKey []byte, prvKey []byte) *Mix {
	return &Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
}

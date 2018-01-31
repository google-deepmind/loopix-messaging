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
	PubKey []byte
	PrvKey []byte
}

func (m *Mix) ProcessPacket(packet []byte, c chan<- []byte, cAdr chan <- sphinx.Hop, cFlag chan <- string, errCh chan <- error){

	nextHop, commands, newPacket, err := sphinx.ProcessSphinxPacket(packet, m.PrvKey)
	if err != nil {
		errCh <- err
	}

	timeoutCh := make(chan []byte, 1)

	go func(p []byte, delay float64) {
		time.Sleep(time.Second * time.Duration(delay))
		timeoutCh <- p
	}(newPacket, commands.Delay)

	c <- <-timeoutCh
	cAdr <- nextHop
	cFlag <- commands.Flag
	errCh <- nil

}

func (m *Mix) SendLoopMessage() {
	// TO DO: this function is currently not used
	fmt.Println("> Sending loop message")
}

func NewMix(id string, pubKey []byte, prvKey []byte) *Mix {
	return &Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
}

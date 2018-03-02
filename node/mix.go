/*
	Package node implements the core functions for a mix node, which allow to process the received cryptographic packets.
*/
package node

import (
	"anonymous-messaging/sphinx"
	"time"
)

// MixNode is the interface that wraps th bacis methods
// of a mix node.
// ProcessPacket processes the received packet bytes and writes in return meta information to given channels.
// GetPublicKey returns the sequence of key bytes.
type MixNode interface {
	ProcessPacket(packet []byte, c chan<- []byte, cAdr chan<- sphinx.Hop, cFlag chan<- string, errCh chan<- error)
	GetPublicKey() []byte
}

type Mix struct {
	pubKey []byte
	prvKey []byte
}

// ProcessPacket performs the processing operation on the received packet, including cryptographic operations and
// extraction of the meta information.
func (m *Mix) ProcessPacket(packet []byte, c chan<- []byte, cAdr chan<- sphinx.Hop, cFlag chan<- string, errCh chan<- error) {

	nextHop, commands, newPacket, err := sphinx.ProcessSphinxPacket(packet, m.prvKey)
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

// GetPublicKey returns the public key of the mixnode.
func (m *Mix) GetPublicKey() []byte {
	return m.pubKey
}

// NewMix creates a new instance of Mix struct with given public and private key
func NewMix(pubKey []byte, prvKey []byte) *Mix {
	return &Mix{pubKey: pubKey, prvKey: prvKey}
}

package node

import (
	"os"
	"testing"

	"anonymous-messaging/publics"
	"github.com/stretchr/testify/assert"
	sphinx "anonymous-messaging/new_packet_format"
)

var mixWorker Mix
var packet sphinx.SphinxPacket
var mixPubs []publics.MixPubs

func TestMain(m *testing.M) {
	pubM, privM := publics.GenerateKeyPair()

	mixWorker = *NewMix("MixWorker", pubM, privM)

	pub1, _ := publics.GenerateKeyPair()
	pub2, _ := publics.GenerateKeyPair()

	m1 := publics.MixPubs{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := publics.MixPubs{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}

	mixPubs = []publics.MixPubs{m1, m2}
	// delays := []float64{1.4, 2.5, 2.3}
	// path := mixPubs

	packet = sphinx.SphinxPacket{}
	os.Exit(m.Run())
}

func TestMixProcessPacket(t *testing.T) {
	ch := make(chan sphinx.SphinxPacket, 1)
	chHop := make(chan sphinx.Hop, 1)
	mixWorker.ProcessPacket(packet, ch, chHop)
	dePacket := <-ch
	nextHop := <- chHop

	expectedPacket := sphinx.SphinxPacket{}
	assert.Equal(t, expectedPacket, dePacket, "Expected to be the same")
	assert.Equal(t, sphinx.Hop{}, nextHop, "Next hope does not match")
}

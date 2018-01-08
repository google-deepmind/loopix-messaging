package node

import (
	"os"
	"testing"

	"anonymous-messaging/packet_format"
	"anonymous-messaging/publics"
	"github.com/stretchr/testify/assert"
)

var mixWorker Mix
var packet packet_format.Packet
var mixPubs []publics.MixPubs

func TestMain(m *testing.M) {
	mixWorker = *NewMix("MixWorker", 0, 0)
	m1 := publics.MixPubs{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: 0}
	m2 := publics.MixPubs{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: 0}

	mixPubs = []publics.MixPubs{m1, m2}
	delays := []float64{1.4, 2.5, 2.3}
	path := mixPubs

	steps := map[string]packet_format.Header{}
	meta1 := packet_format.MetaData{NextHopId: "Mix2", NextHopHost: "localhost", NextHopPort: "3331", FinalFlag: true}
	steps["Mix1"] = packet_format.Header{Meta: meta1, Delay: 1.4}
	packet = packet_format.NewPacket("Hello you", delays, path, steps)
	os.Exit(m.Run())
}

func TestMixProcessPacket(t *testing.T) {
	ch := make(chan packet_format.Packet, 1)
	mixWorker.ProcessPacket(packet, ch)
	dePacket := <-ch

	steps := map[string]packet_format.Header{}
	meta1 := packet_format.MetaData{NextHopId: "Mix2", NextHopHost: "localhost", NextHopPort: "3331", FinalFlag: true}
	steps["Mix1"] = packet_format.Header{Meta: meta1, Delay: 1.4}

	expectedPacket := packet_format.Packet{Message: "Hello you", Path: mixPubs, Delays: []float64{1.4, 2.5, 2.3}, Steps: steps}
	assert.Equal(t, expectedPacket, dePacket, "Expected to be the same")
}

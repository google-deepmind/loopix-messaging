package anonymous_messaging

import (
	"testing"
	"anonymous-messaging/packet_format"
	"github.com/stretchr/testify/assert"
)

func TestMixProcessPacket(t *testing.T) {
	ch := make(chan packet_format.Packet, 1)
	mixWorker.ProcessPacket(packet, ch)
	dePacket := <- ch

	steps := map[string]packet_format.Header{}
	meta1 := packet_format.MetaData{NextHopId:"Mix2", NextHopHost:"localhost", NextHopPort:"3331", FinalFlag:true}
	steps["Mix1"] = packet_format.Header{Meta:meta1, Delay:1.4}

	expectedPacket := packet_format.Packet{Message:"Hello you", Path:mixPubs, Delays:[]float64{1.4, 2.5, 2.3}, Steps:steps}
	assert.Equal(t, expectedPacket, dePacket, "Expected to be the same")
}

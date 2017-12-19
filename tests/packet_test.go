package tests

import (
	"testing"
	"anonymous-messaging/packet_format"
	"github.com/stretchr/testify/assert"
)

func TestPacketToFromString(t *testing.T){
	s := packet_format.ToString(packet)
	expected := packet_format.FromString(s)
	assert.Equal(t, expected, packet, "Conversion to and from string should give the same result")
}

func TestPacketEncode(t *testing.T){
	encoded := packet_format.Encode("Hello you", mixPubs, []float64{1.4, 2.5, 2.3})
	assert.Equal(t, packet, encoded, "Expected to be the same")
}

func TestPacketDecode(t *testing.T){
	decoded := packet_format.Decode(packet)
	expected := packet
	assert.Equal(t, decoded, expected, "The expected and decoded should be the same")
}

package tests

import (
	"testing"
	node "anonymous-messaging/node"
	client "anonymous-messaging/client"
	mix "anonymous-messaging/server"
	packet_format "anonymous-messaging/packet_format"
	"github.com/stretchr/testify/assert"
	"os"
	"anonymous-messaging/publics"
	"reflect"
)

var c client.Client
var mixServer mix.MixServer
var mixWorker node.Mix
var packet packet_format.Packet
var mixPubs []publics.MixPubs

func TestMain(m *testing.M) {
	c = client.NewClient("Client1", "127.0.0.1", "9999", 0, 0)
	mixWorker = node.NewMix("MixWorker", 0, 0)
	m1 := publics.MixPubs{"Mix1", "localhost", "3330", 0}
	m2 := publics.MixPubs{"Mix2", "localhost", "3331", 0}

	mixPubs = []publics.MixPubs{m1, m2}
	delays := []float64{1.4, 2.5, 2.3}
	path := mixPubs

	steps := map[string]packet_format.Header{}
	meta1 := packet_format.MetaData{NextHopId:"Mix2", NextHopHost:"localhost", NextHopPort:"3331", FinalFlag:true}
	steps["Mix1"] = packet_format.Header{Meta:meta1, Delay:1.4}
	packet = packet_format.NewPacket("Hello you", delays, path, steps)
	code := m.Run()
	os.Exit(code)
}

func TestClientEncode(t *testing.T) {

	p := c.EncodeMessage("Hello world", mixPubs, []float64{1.4, 2.5, 2.3})
	assert.Equal(t, p.Message, "Hello world", "The messages should be the same")
	assert.Equal(t, p.Path, mixPubs, "The path should be the same")
}

func TestClientDecode(t *testing.T) {
	d := c.DecodeMessage(packet)
	if reflect.DeepEqual(d, packet) == false {
		t.Error("Error in decode: ")
	}
}

func TestClientSendMessage(t *testing.T) {
}

func TestClientGenerateDelays(t *testing.T){
	delays := c.GenerateDelaySequence(5,3)
	if len(delays) != 3 {
		t.Error("Incorrect number of generated delays")
	}
	if reflect.TypeOf(delays).Elem().Kind() != reflect.Float64 {
		t.Error("Incorrect type of generated delays")
	}
}

func TestClientProcessPacket(t *testing.T){
	m := c.ProcessPacket(packet)
	assert.Equal(t, m, "Hello you", "The final message should be the same as the init one")
}

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



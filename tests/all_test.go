package tests

import (
	"testing"
	node "anonymous-messaging/node"
	client "anonymous-messaging/client"
	mix "anonymous-messaging/server"
	packet_format "anonymous-messaging/packet_format"
	"os"
	"net"
	"fmt"
	"reflect"
)

var c client.Client
var mixServer mix.MixServer
var mixWorker node.Mix
var packet packet_format.Packet


func TestMain(m *testing.M) {
	c = client.NewClient("Client1", "127.0.0.1", "9999", 0, 0)
	mixWorker = node.NewMix("MixWorker", 0, 0)
	delays := []float64{1.4, 2.5, 2.3}
	path := []string{"A", "B", "C"}
	packet = packet_format.NewPacket("Hello you", delays, path)
	code := m.Run()
	os.Exit(code)
}

func TestClientEncode(t *testing.T) {
	p := c.EncodeMessage("Hello world", nil, nil)
	expPacket := packet_format.Packet{"Hello world", nil, nil}
	if reflect.DeepEqual(expPacket, p) == false {
		t.Error("Test for Client Encode failed")
	}
}

func TestClientDecode(t *testing.T) {
	d := c.DecodeMessage("Encoded packet")
	if d != "Encoded packet" {
		t.Error("Error in decode: ")
	}
}

func TestClientSendMessage(t *testing.T) {
}

func TestClientGenerateDelays(t *testing.T){
	delays := c.GenerateDelaySequence(3)
	if len(delays) != 3 {
		t.Error("Incorrect number of generated delays")
	}
	if reflect.TypeOf(delays).Elem().Kind() != reflect.Float64 {
		t.Error("Incorrect type of generated delays")
	}
}

func TestPacketToString(t *testing.T){
	s := packet.ToString()
	expected := packet_format.FromString(s)
	if reflect.DeepEqual(packet, expected) == false {
		t.Error("Incorrect encoding packet to string")
	}
}

func TestMixProcessPacket(t *testing.T) {
	ch := make(chan string, 1)
	mixWorker.ProcessPacket(packet.ToString(), ch)
	dePacket := packet_format.FromString(<- ch)
	expectedPacket := packet_format.Packet{"Hello you", []string{"A", "B", "C"}, []float64{1.4, 2.5, 2.3}}
	if reflect.DeepEqual(dePacket, expectedPacket) == false {
		t.Error("Test for Mix Process Packet failed")
	}
}

func TestConnection(t *testing.T) {
	s, c := net.Pipe()
	fmt.Println(s)
	fmt.Println(c)
	// to DO
}




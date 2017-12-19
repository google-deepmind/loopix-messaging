package tests

import (
	"testing"
	"anonymous-messaging/publics"
	"anonymous-messaging/packet_format"
	node "anonymous-messaging/node"
	client "anonymous-messaging/client"
	mixserver "anonymous-messaging/server"
	"os"
)

var c client.Client
var mixServer mixserver.MixServer
var mixWorker node.Mix
var packet packet_format.Packet
var mixPubs []publics.MixPubs

func TestMain(m *testing.M) {
	c = client.NewClient("Client1", "127.0.0.1", "9999", "./database.db", 0, 0)
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

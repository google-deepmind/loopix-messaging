package packet_format

import (
	"encoding/json"
	"os"
	"testing"

	"anonymous-messaging/publics"
	"github.com/stretchr/testify/assert"
)

var packet Packet
var mixPubs []publics.MixPubs
var recipient publics.MixPubs

func TestMain(m *testing.M) {
	m1 := publics.MixPubs{Id:"Mix1", Host: "localhost", Port: "3330", PubKey: 0}
	m2 := publics.MixPubs{Id:"Mix2", Host: "localhost", Port: "3331", PubKey: 0}
	mixPubs = []publics.MixPubs{m1, m2}

	recipient = publics.MixPubs{Id: "Recipient", Host: "127.0.0.1", Port: "9999", PubKey: 0}

	mixPubs = []publics.MixPubs{m1, m2}

	os.Exit(m.Run())
}

func TestPacketEncode(t *testing.T) {
	path := append(mixPubs, recipient)
	encoded := Encode("Hello you", path, []float64{1.4, 2.5, 2.3})

	header1 := Header{Meta: MetaData{NextHopId: "Mix2", NextHopHost: "localhost", NextHopPort: "3331", StoreFlag: false, FinalFlag: true}, Delay: 1.4}
	header2 := Header{Meta: MetaData{NextHopId: "Recipient", NextHopHost: "127.0.0.1", NextHopPort: "9999", StoreFlag: false, FinalFlag: true}, Delay: 2.5}
	header3 := Header{Meta: MetaData{NextHopId: "", NextHopHost: "", NextHopPort: "", StoreFlag: false, FinalFlag: false}, Delay: 2.3}
	expected := Packet{Message: "Hello you", Path: path, Delays: []float64{1.4, 2.5, 2.3}, Steps: map[string]Header{"Mix1": header1, "Mix2": header2, "Recipient": header3}}

	assert.Equal(t, expected, encoded, "Expected to be the same")
}

func TestPacketDecode(t *testing.T) {
	decoded := Decode(packet)
	expected := packet
	assert.Equal(t, decoded, expected, "The expected and decoded should be so far the same")
}

func TestPacketToString(t *testing.T) {
	s := ToString(packet)
	expected, err := json.Marshal(packet)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, s, string(expected), "Conversion to string should give the same result")
}

func TestPacketFromString(t *testing.T) {
	asString, err := json.Marshal(packet)
	if err != nil {
		panic(err)
	}
	expected := packet
	fromString := FromString(string(asString))
	assert.Equal(t, fromString, expected, "Conversion from string should give the same result")
}

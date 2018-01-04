package packet_format

import (
	"anonymous-messaging/publics"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var packet Packet
var mixPubs []publics.MixPubs
var recipient publics.MixPubs

func TestMain(m *testing.M) {
	m1 := publics.MixPubs{"Mix1", "localhost", "3330", 0}
	m2 := publics.MixPubs{"Mix2", "localhost", "3331", 0}
	mixPubs = []publics.MixPubs{m1, m2}

	recipient = publics.MixPubs{"Recipient", "127.0.0.1", "9999", 0}

	mixPubs = []publics.MixPubs{m1, m2}

	code := m.Run()
	os.Exit(code)
}

func TestPacketEncode(t *testing.T) {
	path := append(mixPubs, recipient)
	encoded := Encode("Hello you", path, []float64{1.4, 2.5, 2.3})

	header1 := Header{MetaData{"Mix2", "localhost", "3331", false, true}, 1.4}
	header2 := Header{MetaData{"Recipient", "127.0.0.1", "9999", false, true}, 2.5}
	header3 := Header{MetaData{"", "", "", false, false}, 2.3}
	expected := Packet{"Hello you", path, []float64{1.4, 2.5, 2.3}, map[string]Header{"Mix1": header1, "Mix2": header2, "Recipient": header3}}

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

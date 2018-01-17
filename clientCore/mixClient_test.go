package clientCore

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"anonymous-messaging/publics"
	"github.com/stretchr/testify/assert"
	sphinx "anonymous-messaging/new_packet_format"
	"crypto/elliptic"
)

var cryptoClient CryptoClient
var mixPubs []publics.MixPubs

func TestMain(m *testing.M) {
	pubC, privC := publics.GenerateKeyPair()
	pub1, _ := publics.GenerateKeyPair()
	pub2, _ := publics.GenerateKeyPair()

	cryptoClient = CryptoClient{Id: "MixClient", PubKey: pubC, PrvKey: privC, Curve: elliptic.P224()}

	m1 := publics.MixPubs{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := publics.MixPubs{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}
	mixPubs = []publics.MixPubs{m1, m2}

	os.Exit(m.Run())
}

func TestMixClientEncode(t *testing.T) {

	message := "Hello world"
	path := mixPubs
	delays := []float64{1.4, 2.5, 2.3}

	pubD, _ := publics.GenerateKeyPair()
	recipient := publics.MixPubs{Id: "Recipient", Host: "localhost", Port: "9999", PubKey: pubD}

	var pubs []publics.PublicKey
	var commands []sphinx.Commands

	for _, v := range path {
		pubs = append(pubs, v.PubKey)
	}

	for _, v := range delays {
		c := sphinx.Commands{Delay: v, Flag: "Flag"}
		commands = append(commands, c)
	}
	encoded := cryptoClient.EncodeMessage(message, path, delays, recipient)
	fmt.Println(encoded)

}

func TestMixClientDecode(t *testing.T) {
	packet := sphinx.SphinxPacket{Hdr: sphinx.Header{}, Pld: []byte("Message")}


	decoded := cryptoClient.DecodeMessage(packet)
	expected := packet

	assert.Equal(t,expected, decoded)
}

func TestGenerateDelaySequence(t *testing.T) {
	delays := cryptoClient.GenerateDelaySequence(100, 5)
	if len(delays) != 5 {
		t.Error("Wrong length")
	}
	if reflect.TypeOf(delays).Elem().Kind() != reflect.Float64 {
		t.Error("Incorrect type of generated delays")
	}
}

func TestGetRandomMixSequence(t *testing.T) {
	// test two cases: the one when len is smaller than all mixes and the one when length is larger / the same
	var mixes []publics.MixPubs
	for i := 0; i < 5; i++ {
		pub, _ := publics.GenerateKeyPair()
		mixes = append(mixes, publics.NewMixPubs(fmt.Sprintf("Mix%d", i), "localhost", strconv.Itoa(3330+i), pub))
	}

	var sequence []publics.MixPubs
	sequence = cryptoClient.GetRandomMixSequence(mixes, 6)
	assert.Equal(t, 5, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")

	sequence = cryptoClient.GetRandomMixSequence(mixes, 3)
	assert.Equal(t, 3, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")
}

package clientCore

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"anonymous-messaging/publics"
	"github.com/stretchr/testify/assert"
	sphinx "anonymous-messaging/sphinx"
	"crypto/elliptic"
)

var cryptoClient CryptoClient
var mixPubs []publics.MixPubs

func TestMain(m *testing.M) {
	pubC, privC := sphinx.GenerateKeyPair()
	pub1, _ := sphinx.GenerateKeyPair()
	pub2, _ := sphinx.GenerateKeyPair()

	cryptoClient = CryptoClient{Id: "MixClient", PubKey: pubC, PrvKey: privC, Curve: elliptic.P224()}

	m1 := publics.MixPubs{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := publics.MixPubs{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}
	mixPubs = []publics.MixPubs{m1, m2}

	os.Exit(m.Run())
}

func TestMixClientEncode(t *testing.T) {

	message := "Hello world"
	delays := []float64{1.4, 2.5, 2.3}

	pubD, _ := sphinx.GenerateKeyPair()
	recipient := publics.MixPubs{Id: "Recipient", Host: "localhost", Port: "9999", PubKey: pubD}

	var pubs [][]byte
	for _, v := range mixPubs {
		pubs = append(pubs, v.PubKey)
	}

	var commands []sphinx.Commands
	for _, v := range delays {
		c := sphinx.Commands{Delay: v, Flag: "Flag"}
		commands = append(commands, c)
	}
	encoded := cryptoClient.EncodeMessage(message, mixPubs, delays, recipient)

	assert.Equal(t, reflect.TypeOf([]byte{}), reflect.TypeOf(encoded))

}

func TestMixClientDecode(t *testing.T) {
	packet := sphinx.SphinxPacket{Hdr: &sphinx.Header{}, Pld: []byte("Message")}

	decoded := cryptoClient.DecodeMessage(packet)
	expected := packet

	assert.Equal(t, expected, decoded)
}

func TestGenerateDelaySequence(t *testing.T) {
	delays := cryptoClient.GenerateDelaySequence(100, 5)

	assert.Equal(t, len(delays), 5, "The length of returned delays should be equal to theinput length")
	assert.Equal(t, reflect.TypeOf([]float64{}), reflect.TypeOf(delays), "The delays should be in float64 type")
}

func TestGetRandomMixSequence(t *testing.T) {
	var mixes []publics.MixPubs
	for i := 0; i < 5; i++ {
		pub, _ := sphinx.GenerateKeyPair()
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

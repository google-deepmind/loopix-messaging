package clientCore

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"anonymous-messaging/config"
	"github.com/stretchr/testify/assert"
	sphinx "anonymous-messaging/sphinx"
	"crypto/elliptic"
)

var cryptoClient CryptoClient
var path config.E2EPath

func TestMain(m *testing.M) {
	pubC, privC, err := sphinx.GenerateKeyPair()
	pub1, _, _ := sphinx.GenerateKeyPair()
	pub2, _, _ := sphinx.GenerateKeyPair()
	pubP, _, _ := sphinx.GenerateKeyPair()
	pubD, _, _ := sphinx.GenerateKeyPair()
	if err != nil{
		panic(err)
	}

	cryptoClient = CryptoClient{Id: "MixClient", PubKey: pubC, PrvKey: privC, Curve: elliptic.P224()}

	m1 := config.MixPubs{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := config.MixPubs{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}
	provider := config.MixPubs{Id: "Provider", Host: "localhost", Port: "3331", PubKey: pubP}
	recipient := config.ClientPubs{Id: "Recipient", Host: "localhost", Port: "9999", PubKey: pubD, Provider: &provider}

	path = config.E2EPath{IngressProvider: provider, Mixes: []config.MixPubs{m1, m2}, EgressProvider: provider, Recipient: recipient}

	os.Exit(m.Run())
}

func TestMixClientEncode(t *testing.T) {

	message := "Hello world"
	delays := []float64{1.4, 2.5, 2.3, 3.5, 6.7}

	var commands []sphinx.Commands
	for _, v := range delays {
		c := sphinx.Commands{Delay: v, Flag: "Flag"}
		commands = append(commands, c)
	}
	encoded, err := cryptoClient.EncodeMessage(message, path, delays)
	if err != nil{
		t.Error(err)
	}

	assert.Equal(t, reflect.TypeOf([]byte{}), reflect.TypeOf(encoded))

}

func TestMixClientDecode(t *testing.T) {
	packet := sphinx.SphinxPacket{Hdr: &sphinx.Header{}, Pld: []byte("Message")}

	decoded, err := cryptoClient.DecodeMessage(packet)
	if err != nil{
		t.Error(err)
	}
	expected := packet

	assert.Equal(t, expected, decoded)
}

func TestGenerateDelaySequence(t *testing.T) {
	delays := cryptoClient.GenerateDelaySequence(100, 5)

	assert.Equal(t, len(delays), 5, "The length of returned delays should be equal to theinput length")
	assert.Equal(t, reflect.TypeOf([]float64{}), reflect.TypeOf(delays), "The delays should be in float64 type")
}

func TestGetRandomMixSequence(t *testing.T) {
	var mixes []config.MixPubs
	for i := 0; i < 5; i++ {
		pub, _, err := sphinx.GenerateKeyPair()
		if err != nil{
			t.Error(err)
		}
		mixes = append(mixes, config.NewMixPubs(fmt.Sprintf("Mix%d", i), "localhost", strconv.Itoa(3330+i), pub))
	}

	var sequence []config.MixPubs
	sequence = cryptoClient.GetRandomMixSequence(mixes, 6)
	assert.Equal(t, 5, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")

	sequence = cryptoClient.GetRandomMixSequence(mixes, 3)
	assert.Equal(t, 3, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")
}

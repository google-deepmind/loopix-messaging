package node

import (
	"os"
	"testing"

	"anonymous-messaging/config"
	sphinx "anonymous-messaging/sphinx"
	"crypto/elliptic"
	"github.com/stretchr/testify/assert"
	"reflect"
)

var providerWorker Mix
var testPacket sphinx.SphinxPacket
var nodes []config.MixPubs
var curve elliptic.Curve

func TestMain(m *testing.M) {
	curve := elliptic.P224()

	pub1, _, _ := sphinx.GenerateKeyPair()
	pub2, _, _ := sphinx.GenerateKeyPair()
	pub3, _, _ := sphinx.GenerateKeyPair()

	pubP, privP, _ := sphinx.GenerateKeyPair()

	m1 := config.MixPubs{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := config.MixPubs{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}
	m3 := config.MixPubs{Id: "Mix2", Host: "localhost", Port: "3332", PubKey: pub3}
	provider := config.MixPubs{Id: "Provider", Host: "localhost", Port: "3333", PubKey: pubP}

	providerWorker = *NewMix("ProviderWorker", pubP, privP)

	nodes = []config.MixPubs{m1, m2, m3}

	pubD, _, _ := sphinx.GenerateKeyPair()
	dest := config.ClientPubs{Id : "Destination", Host: "localhost", Port: "3334", PubKey: pubD, Provider: &provider}

	path := config.E2EPath{IngressProvider: provider, Mixes: []config.MixPubs{m1, m2, m3}, EgressProvider: provider, Recipient: dest}

	var err error
	testPacket, err = sphinx.PackForwardMessage(curve, path, []float64{1.4, 2.5, 2.3, 3.2, 7.4}, "Test Message")
	if err != nil{
		panic(err)
	}
	os.Exit(m.Run())
}

func TestMixProcessPacket(t *testing.T) {
	ch := make(chan []byte, 1)
	chHop := make(chan string, 1)
	cAdr := make(chan string, 1)

	testPacketBytes, err := testPacket.Bytes()
	if err != nil{
		t.Error(err)
	}

	providerWorker.ProcessPacket(testPacketBytes, ch, chHop, cAdr)
	dePacket := <-ch
	nextHop := <- chHop
	flag := <- cAdr

	assert.Equal(t, "localhost:3330", nextHop, "Next hope does not match")
	assert.Equal(t, reflect.TypeOf([]byte{}), reflect.TypeOf(dePacket))
	assert.Equal(t, "\xF1", flag, reflect.TypeOf(dePacket))
}

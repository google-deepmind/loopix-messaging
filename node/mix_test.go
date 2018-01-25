package node

import (
	"os"
	"testing"

	"anonymous-messaging/publics"
	sphinx "anonymous-messaging/sphinx"
	"crypto/elliptic"
	"github.com/stretchr/testify/assert"
	"reflect"
)

var providerWorker Mix
var testPacket sphinx.SphinxPacket
var nodes []publics.MixPubs
var curve elliptic.Curve

func TestMain(m *testing.M) {
	curve := elliptic.P224()

	pub1, _ := sphinx.GenerateKeyPair()
	pub2, _ := sphinx.GenerateKeyPair()
	pub3, _ := sphinx.GenerateKeyPair()

	pubP, privP := sphinx.GenerateKeyPair()

	m1 := publics.MixPubs{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := publics.MixPubs{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}
	m3 := publics.MixPubs{Id: "Mix2", Host: "localhost", Port: "3332", PubKey: pub3}
	provider := publics.MixPubs{Id: "Provider", Host: "localhost", Port: "3333", PubKey: pubP}

	providerWorker = *NewMix("ProviderWorker", pubP, privP)

	nodes = []publics.MixPubs{m1, m2, m3}

	pubD, _ := sphinx.GenerateKeyPair()
	dest := publics.ClientPubs{Id : "Destination", Host: "localhost", Port: "3334", PubKey: pubD, Provider: &provider}

	path := publics.E2EPath{IngressProvider: provider, Mixes: []publics.MixPubs{m1, m2, m3}, EgressProvider: provider, Recipient: dest}

	testPacket = sphinx.PackForwardMessage(curve, path, []float64{1.4, 2.5, 2.3, 3.2, 7.4}, "Test Message")
	os.Exit(m.Run())
}

func TestMixProcessPacket(t *testing.T) {
	ch := make(chan []byte, 1)
	chHop := make(chan string, 1)
	cAdr := make(chan string, 1)

	providerWorker.ProcessPacket(testPacket.Bytes(), ch, chHop, cAdr)
	dePacket := <-ch
	nextHop := <- chHop
	flag := <- cAdr

	assert.Equal(t, "localhost:3330", nextHop, "Next hope does not match")
	assert.Equal(t, reflect.TypeOf([]byte{}), reflect.TypeOf(dePacket))
	assert.Equal(t, "\xF1", flag, reflect.TypeOf(dePacket))
}

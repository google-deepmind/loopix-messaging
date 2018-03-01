package node

import (
	"anonymous-messaging/config"
	sphinx "anonymous-messaging/sphinx"

	"github.com/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"crypto/elliptic"
	"os"
	"reflect"
	"testing"
)

var providerWorker *Mix
var testPacket *sphinx.SphinxPacket
var nodes []config.MixConfig
var curve elliptic.Curve

func Setup() error {
	curve := elliptic.P224()

	pub1, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}
	pub2, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}
	pub3, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}
	pubP, privP, _ := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}

	m1 := config.MixConfig{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := config.MixConfig{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}
	m3 := config.MixConfig{Id: "Mix2", Host: "localhost", Port: "3332", PubKey: pub3}
	provider := config.MixConfig{Id: "Provider", Host: "localhost", Port: "3333", PubKey: pubP}

	providerWorker = NewMix(pubP, privP)
	nodes = []config.MixConfig{m1, m2, m3}

	pubD, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}

	dest := config.ClientConfig{Id: "Destination", Host: "localhost", Port: "3334", PubKey: pubD, Provider: &provider}
	path := config.E2EPath{IngressProvider: provider, Mixes: []config.MixConfig{m1, m2, m3}, EgressProvider: provider, Recipient: dest}

	*testPacket, err = sphinx.PackForwardMessage(curve, path, []float64{1.4, 2.5, 2.3, 3.2, 7.4}, "Test Message")
	if err != nil {
		panic(err)
	}

	return nil
}

func TestMain(m *testing.M) {

	err := Setup()
	if err != nil {
		panic(m)
	}
	os.Exit(m.Run())
}

func TestMixProcessPacket(t *testing.T) {
	ch := make(chan []byte, 1)
	chHop := make(chan sphinx.Hop, 1)
	cAdr := make(chan string, 1)
	errCh := make(chan error, 1)

	testPacketBytes, err := proto.Marshal(testPacket)
	if err != nil {
		t.Fatal(err)
	}

	providerWorker.ProcessPacket(testPacketBytes, ch, chHop, cAdr, errCh)
	dePacket := <-ch
	nextHop := <-chHop
	flag := <-cAdr
	err = <-errCh
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, sphinx.Hop{Id: "Mix1", Address: "localhost:3330", PubKey: nodes[0].PubKey}, nextHop, "Next hop does not match")
	assert.Equal(t, reflect.TypeOf([]byte{}), reflect.TypeOf(dePacket))
	assert.Equal(t, "\xF1", flag, reflect.TypeOf(dePacket))
}

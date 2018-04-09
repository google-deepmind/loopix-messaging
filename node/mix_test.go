// Copyright 2018 The Loopix-Messaging Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

var nodes []config.MixConfig

func createProviderWorker() (*Mix, error) {
	pubP, privP, err := sphinx.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	providerWorker := NewMix(pubP, privP)
	return providerWorker, nil
}

func createTestPacket(curve elliptic.Curve, mixes []config.MixConfig, provider config.MixConfig, recipient config.ClientConfig) (*sphinx.SphinxPacket, error) {
	path := config.E2EPath{IngressProvider: provider, Mixes: mixes, EgressProvider: provider, Recipient: recipient}
	testPacket, err := sphinx.PackForwardMessage(curve, path, []float64{1.4, 2.5, 2.3, 3.2, 7.4}, "Test Message")
	if err != nil {
		return nil, err
	}
	return &testPacket, nil
}

func createTestMixes() ([]config.MixConfig, error) {
	pub1, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	pub2, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	pub3, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	m1 := config.MixConfig{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := config.MixConfig{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}
	m3 := config.MixConfig{Id: "Mix2", Host: "localhost", Port: "3332", PubKey: pub3}
	nodes = []config.MixConfig{m1, m2, m3}

	return nodes, nil
}

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

func TestMixProcessPacket(t *testing.T) {
	ch := make(chan []byte, 1)
	chHop := make(chan sphinx.Hop, 1)
	cAdr := make(chan string, 1)
	errCh := make(chan error, 1)

	pubD, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	providerWorker, err := createProviderWorker()
	if err != nil {
		t.Fatal(err)
	}
	provider := config.MixConfig{Id: "Provider", Host: "localhost", Port: "3333", PubKey: providerWorker.pubKey}
	dest := config.ClientConfig{Id: "Destination", Host: "localhost", Port: "3334", PubKey: pubD, Provider: &provider}
	mixes, err := createTestMixes()
	if err != nil {
		t.Fatal(err)
	}

	testPacket, err := createTestPacket(elliptic.P224(), mixes, provider, dest)
	if err != nil {
		t.Fatal(err)
	}

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

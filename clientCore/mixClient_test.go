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

package clientCore

import (
	"anonymous-messaging/config"
	sphinx "anonymous-messaging/sphinx"

	"github.com/stretchr/testify/assert"

	"crypto/elliptic"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
)

var client *CryptoClient
var path config.E2EPath
var mixes []config.MixConfig

func Setup() error {
	for i := 0; i < 10; i++ {
		pub, _, err := sphinx.GenerateKeyPair()
		if err != nil {
			return err
		}
		mixes = append(mixes, config.NewMixConfig(fmt.Sprintf("Mix%d", i), "localhost", strconv.Itoa(3330+i), pub))
	}

	// Create a mixClient
	pubC, privC, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}
	client = NewCryptoClient(pubC, privC, elliptic.P224(), config.MixConfig{}, NetworkPKI{})

	//Client a pair of mix configs, a single provider and a recipient
	pub1, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}

	pub2, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		return err
	}

	m1 := config.MixConfig{Id: "Mix1", Host: "localhost", Port: "3330", PubKey: pub1}
	m2 := config.MixConfig{Id: "Mix2", Host: "localhost", Port: "3331", PubKey: pub2}

	client.Network = NetworkPKI{}
	client.Network.Mixes = []config.MixConfig{m1, m2}

	return nil
}

func TestMain(m *testing.M) {

	err := Setup()
	if err != nil {
		panic(m)

	}
	os.Exit(m.Run())
}

func TestCryptoClient_EncodeMessage(t *testing.T) {

	pubP, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	provider := config.MixConfig{Id: "Provider", Host: "localhost", Port: "3331", PubKey: pubP}

	pubD, _, err := sphinx.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	recipient := config.ClientConfig{Id: "Recipient", Host: "localhost", Port: "9999", PubKey: pubD, Provider: &provider}
	client.Provider = provider

	encoded, err := client.EncodeMessage("Hello world", recipient)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, reflect.TypeOf([]byte{}), reflect.TypeOf(encoded))

}

func TestCryptoClient_DecodeMessage(t *testing.T) {
	packet := sphinx.SphinxPacket{Hdr: &sphinx.Header{}, Pld: []byte("Message")}

	decoded, err := client.DecodeMessage(packet)
	if err != nil {
		t.Fatal(err)
	}
	expected := packet
	assert.Equal(t, expected, decoded)
}

func TestCryptoClient_GenerateDelaySequence_Pass(t *testing.T) {
	delays, err := client.generateDelaySequence(100, 5)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(delays), 5, "The length of returned delays should be equal to theinput length")
	assert.Equal(t, reflect.TypeOf([]float64{}), reflect.TypeOf(delays), "The delays should be in float64 type")
}

func TestCryptoClient_GenerateDelaySequence_Fail(t *testing.T) {
	_, err := client.generateDelaySequence(0, 5)
	assert.EqualError(t, errors.New("the parameter of exponential distribution has to be larger than zero"), err.Error(), "")
}

func Test_GetRandomMixSequence_TooFewMixes(t *testing.T) {

	sequence, err := client.getRandomMixSequence(mixes, 20)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")
}

func Test_GetRandomMixSequence_MoreMixes(t *testing.T) {

	sequence, err := client.getRandomMixSequence(mixes, 3)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(sequence), "When the given length is larger than the number of active nodes, the path should be "+
		"the sequence of all active mixes")

}

func Test_GetRandomMixSequence_FailEmptyList(t *testing.T) {
	_, err := client.getRandomMixSequence([]config.MixConfig{}, 6)
	assert.EqualError(t, errors.New("cannot take a mix sequence from an empty list"), err.Error(), "")
}

func Test_GetRandomMixSequence_FailNonList(t *testing.T) {
	_, err := client.getRandomMixSequence(nil, 6)
	assert.EqualError(t, errors.New("cannot take a mix sequence from an empty list"), err.Error(), "")
}

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

package sphinx

import (
	"anonymous-messaging/config"

	"crypto/aes"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"fmt"
	"math/big"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	curve = elliptic.P224()

	os.Exit(m.Run())
}

func TestExpoSingleValue(t *testing.T) {
	_, x, y, err := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil {
		t.Error(err)
	}

	randomPoint := elliptic.Marshal(curve, x, y)
	nBig := *big.NewInt(2)
	exp := []big.Int{nBig}

	result := expo(randomPoint, exp)
	expectedX, expectedY := curve.ScalarMult(x, y, nBig.Bytes())
	assert.Equal(t, elliptic.Marshal(curve, expectedX, expectedY), result)

}

func TestExpoMultipleValue(t *testing.T) {
	_, x, y, err := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil {
		t.Error(err)
	}
	randomPoint := elliptic.Marshal(curve, x, y)

	var exp []big.Int
	for i := 1; i <= 5; i++ {
		exp = append(exp, *big.NewInt(int64(i)))
	}

	result := expo(randomPoint, exp)
	expectedX, expectedY := curve.ScalarMult(x, y, big.NewInt(120).Bytes())
	assert.Equal(t, elliptic.Marshal(curve, expectedX, expectedY), result)
}

func TestExpoBaseSingleValue(t *testing.T) {
	nBig := *big.NewInt(2)
	exp := []big.Int{nBig}

	result := expoGroupBase(curve, exp)
	expectedX, expectedY := curve.ScalarBaseMult(nBig.Bytes())

	assert.Equal(t, elliptic.Marshal(curve, expectedX, expectedY), result)
}

func TestExpoBaseMultipleValue(t *testing.T) {
	var exp []big.Int
	for i := 1; i <= 3; i++ {
		exp = append(exp, *big.NewInt(int64(i)))
	}
	result := expoGroupBase(curve, exp)
	expectedX, expectedY := curve.ScalarBaseMult(big.NewInt(6).Bytes())
	assert.Equal(t, elliptic.Marshal(curve, expectedX, expectedY), result)

}

func TestHash(t *testing.T) {
	_, x, y, err := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil {
		t.Error(err)
	}

	randomPoint := elliptic.Marshal(curve, x, y)
	hVal := hash(randomPoint)

	assert.Equal(t, 32, len(hVal))

}

func TestBytesToBigNum(t *testing.T) {
	bytes := big.NewInt(100).Bytes()
	result := *bytesToBigNum(curve, bytes)
	assert.Equal(t, *big.NewInt(100), result)
}

func TestGetAESKey(t *testing.T) {
	_, x, y, err := elliptic.GenerateKey(curve, rand.Reader)

	if err != nil {
		t.Error(err)
	}

	randomPoint := elliptic.Marshal(curve, x, y)
	aesKey := KDF(randomPoint)
	assert.Equal(t, aes.BlockSize, len(aesKey))

}

func TestComputeBlindingFactor(t *testing.T) {
	generator := elliptic.Marshal(curve, curve.Params().Gx, curve.Params().Gy)

	key := hash(generator)
	b, err := computeBlindingFactor(curve, key)
	if err != nil {
		t.Error(err)
	}

	expected := new(big.Int)
	expected.SetString("252286146058081748716688845275111486959", 10)

	assert.Equal(t, expected, b)
}

func TestGetSharedSecrets(t *testing.T) {

	pub1, _, err := GenerateKeyPair()
	pub2, _, err := GenerateKeyPair()
	pub3, _, err := GenerateKeyPair()
	if err != nil {
		t.Error(err)
	}

	pubs := [][]byte{pub1, pub2, pub3}

	m1 := config.MixConfig{Id: "", Host: "", Port: "", PubKey: pub1}
	m2 := config.MixConfig{Id: "", Host: "", Port: "", PubKey: pub2}
	m3 := config.MixConfig{Id: "", Host: "", Port: "", PubKey: pub3}

	nodes := []config.MixConfig{m1, m2, m3}

	x := big.NewInt(100)

	result, err := getSharedSecrets(curve, nodes, *x)
	if err != nil {
		t.Error(err)
	}

	var expected []HeaderInitials
	blindFactors := []big.Int{*x}

	v := x
	alpha0X, alpha0Y := curve.Params().ScalarMult(curve.Params().Gx, curve.Params().Gy, v.Bytes())
	alpha0 := elliptic.Marshal(curve, alpha0X, alpha0Y)
	s0 := expo(pubs[0], blindFactors)
	aesS0 := KDF(s0)
	b0, err := computeBlindingFactor(curve, aesS0)
	if err != nil {
		t.Error(err)
	}

	expected = append(expected, HeaderInitials{Alpha: alpha0, Secret: s0, Blinder: b0.Bytes(), SecretHash: aesS0})
	blindFactors = append(blindFactors, *b0)

	v = big.NewInt(0).Mul(v, b0)
	alpha1X, alpha1Y := curve.Params().ScalarMult(curve.Params().Gx, curve.Params().Gy, v.Bytes())
	alpha1 := elliptic.Marshal(curve, alpha1X, alpha1Y)
	s1 := expo(pubs[1], blindFactors)
	aesS1 := KDF(s1)
	b1, err := computeBlindingFactor(curve, aesS1)
	if err != nil {
		t.Error(err)
	}

	expected = append(expected, HeaderInitials{Alpha: alpha1, Secret: s1, Blinder: b1.Bytes(), SecretHash: aesS1})
	blindFactors = append(blindFactors, *b1)

	v = big.NewInt(0).Mul(v, b1)
	alpha2X, alpha2Y := curve.Params().ScalarMult(curve.Params().Gx, curve.Params().Gy, v.Bytes())
	alpha2 := elliptic.Marshal(curve, alpha2X, alpha2Y)
	s2 := expo(pubs[2], blindFactors)
	aesS2 := KDF(s2)
	b2, err := computeBlindingFactor(curve, aesS2)
	if err != nil {
		t.Error(err)
	}

	expected = append(expected, HeaderInitials{Alpha: alpha2, Secret: s2, Blinder: b2.Bytes(), SecretHash: aesS2})
	blindFactors = append(blindFactors, *b2)

	assert.Equal(t, expected, result)
}

func TestComputeFillers(t *testing.T) {

	g := elliptic.Marshal(curve, curve.Params().Gx, curve.Params().Gy)
	h1 := HeaderInitials{Alpha: []byte{}, Secret: g, Blinder: []byte{}, SecretHash: []byte("1111111111111111")}
	h2 := HeaderInitials{Alpha: []byte{}, Secret: g, Blinder: []byte{}, SecretHash: []byte("1111111111111111")}
	h3 := HeaderInitials{Alpha: []byte{}, Secret: g, Blinder: []byte{}, SecretHash: []byte("1111111111111111")}
	tuples := []HeaderInitials{h1, h2, h3}

	pub1, _, err := GenerateKeyPair()
	pub2, _, err := GenerateKeyPair()
	pub3, _, err := GenerateKeyPair()
	if err != nil {
		t.Error(err)
	}

	m1 := config.MixConfig{Id: "", Host: "", Port: "", PubKey: pub1}
	m2 := config.MixConfig{Id: "", Host: "", Port: "", PubKey: pub2}
	m3 := config.MixConfig{Id: "", Host: "", Port: "", PubKey: pub3}

	fillers, err := computeFillers([]config.MixConfig{m1, m2, m3}, tuples)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("FILLER: ", fillers)

}

func TestXorBytesPass(t *testing.T) {
	result := XorBytes([]byte("00101"), []byte("10110"))
	assert.Equal(t, []byte{1, 0, 0, 1, 1}, result)
}

func TestXorBytesFail(t *testing.T) {
	result := XorBytes([]byte("00101"), []byte("10110"))
	assert.NotEqual(t, []byte("00000"), result)
}

func TestEncapsulateHeader(t *testing.T) {

	pub1, _, err := GenerateKeyPair()
	pub2, _, err := GenerateKeyPair()
	pub3, _, err := GenerateKeyPair()
	pubD, _, err := GenerateKeyPair()
	if err != nil {
		t.Error(err)
	}

	m1 := config.NewMixConfig("Node1", "localhost", "3331", pub1)
	m2 := config.NewMixConfig("Node2", "localhost", "3332", pub2)
	m3 := config.NewMixConfig("Node3", "localhost", "3333", pub3)

	nodes := []config.MixConfig{m1, m2, m3}

	c1 := Commands{Delay: 0.34, Flag: "0"}
	c2 := Commands{Delay: 0.25, Flag: "1"}
	c3 := Commands{Delay: 1.10, Flag: "1"}
	commands := []Commands{c1, c2, c3}

	x := big.NewInt(100)
	sharedSecrets, err := getSharedSecrets(curve, nodes, *x)
	if err != nil {
		t.Error(err)
	}

	actualHeader, err := encapsulateHeader(sharedSecrets, nodes, commands,
		config.ClientConfig{Id: "DestinationId", Host: "DestinationAddress", Port: "9998", PubKey: pubD})
	if err != nil {
		t.Error(err)
	}

	routing1 := RoutingInfo{NextHop: &Hop{"DestinationId", "DestinationAddress:9998", []byte{}}, RoutingCommands: &c3,
		NextHopMetaData: []byte{}, Mac: []byte{}}

	routing1Bytes, err := proto.Marshal(&routing1)
	if err != nil {
		t.Error(err)
	}

	enc_routing1, err := AES_CTR(KDF(sharedSecrets[2].SecretHash), routing1Bytes)
	if err != nil {
		t.Error(err)
	}

	mac1 := computeMac(KDF(sharedSecrets[2].SecretHash), enc_routing1)

	routing2 := RoutingInfo{NextHop: &Hop{"Node3", "localhost:3333", pub3}, RoutingCommands: &c2,
		NextHopMetaData: enc_routing1, Mac: mac1}

	routing2Bytes, err := proto.Marshal(&routing2)
	if err != nil {
		t.Error(err)
	}

	enc_routing2, err := AES_CTR(KDF(sharedSecrets[1].SecretHash), routing2Bytes)
	if err != nil {
		t.Error(err)
	}

	mac2 := computeMac(KDF(sharedSecrets[1].SecretHash), enc_routing2)

	expectedRouting := RoutingInfo{NextHop: &Hop{"Node2", "localhost:3332", pub2}, RoutingCommands: &c1,
		NextHopMetaData: enc_routing2, Mac: mac2}

	expectedRoutingBytes, err := proto.Marshal(&expectedRouting)
	if err != nil {
		t.Error(err)
	}

	enc_expectedRouting, err := AES_CTR(KDF(sharedSecrets[0].SecretHash), expectedRoutingBytes)
	if err != nil {
		t.Error(err)
	}

	mac3 := computeMac(KDF(sharedSecrets[0].SecretHash), enc_expectedRouting)

	expectedHeader := Header{sharedSecrets[0].Alpha, enc_expectedRouting, mac3}

	assert.Equal(t, expectedHeader, actualHeader)
}

func TestProcessSphinxHeader(t *testing.T) {

	pub1, priv1, err := GenerateKeyPair()
	pub2, _, err := GenerateKeyPair()
	pub3, _, err := GenerateKeyPair()
	if err != nil {
		t.Error(err)
	}

	c1 := Commands{Delay: 0.34}
	c2 := Commands{Delay: 0.25}
	c3 := Commands{Delay: 1.10}

	m1 := config.NewMixConfig("Node1", "localhost", "3331", pub1)
	m2 := config.NewMixConfig("Node2", "localhost", "3332", pub2)
	m3 := config.NewMixConfig("Node3", "localhost", "3333", pub3)

	nodes := []config.MixConfig{m1, m2, m3}

	x := big.NewInt(100)
	sharedSecrets, err := getSharedSecrets(curve, nodes, *x)
	if err != nil {
		t.Error(err)
	}

	// Intermediate steps, which are needed to check whether the processing of the header was correct
	routing1 := RoutingInfo{NextHop: &Hop{"DestinationId", "DestinationAddress", []byte{}}, RoutingCommands: &c3,
		NextHopMetaData: []byte{}, Mac: []byte{}}

	routing1Bytes, err := proto.Marshal(&routing1)
	if err != nil {
		t.Error(err)
	}

	enc_routing1, err := AES_CTR(KDF(sharedSecrets[2].SecretHash), routing1Bytes)
	if err != nil {
		t.Error(err)
	}

	mac1 := computeMac(KDF(sharedSecrets[2].SecretHash), enc_routing1)

	routing2 := RoutingInfo{NextHop: &Hop{"Node3", "localhost:3333", pub3}, RoutingCommands: &c2,
		NextHopMetaData: enc_routing1, Mac: mac1}

	routing2Bytes, err := proto.Marshal(&routing2)
	if err != nil {
		t.Error(err)
	}

	enc_routing2, err := AES_CTR(KDF(sharedSecrets[1].SecretHash), routing2Bytes)
	if err != nil {
		t.Error(err)
	}

	mac2 := computeMac(KDF(sharedSecrets[1].SecretHash), enc_routing2)

	routing3 := RoutingInfo{NextHop: &Hop{"Node2", "localhost:3332", pub2}, RoutingCommands: &c1,
		NextHopMetaData: enc_routing2, Mac: mac2}

	routing3Bytes, err := proto.Marshal(&routing3)
	if err != nil {
		t.Error(err)
	}

	enc_expectedRouting, err := AES_CTR(KDF(sharedSecrets[0].SecretHash), routing3Bytes)
	if err != nil {
		t.Error(err)
	}

	mac3 := computeMac(KDF(sharedSecrets[0].SecretHash), enc_expectedRouting)

	header := Header{sharedSecrets[0].Alpha, enc_expectedRouting, mac3}

	nextHop, newCommands, newHeader, err := ProcessSphinxHeader(header, priv1)

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, nextHop, Hop{Id: "Node2", Address: "localhost:3332", PubKey: pub2})
	assert.Equal(t, newCommands, c1)
	assert.Equal(t, newHeader, Header{Alpha: sharedSecrets[1].Alpha, Beta: enc_routing2, Mac: mac2})

}

func TestProcessSphinxPayload(t *testing.T) {

	message := "Plaintext message"

	pub1, priv1, err := GenerateKeyPair()
	pub2, priv2, err := GenerateKeyPair()
	pub3, priv3, err := GenerateKeyPair()
	if err != nil {
		t.Error(err)
	}

	m1 := config.NewMixConfig("Node1", "localhost", "3331", pub1)
	m2 := config.NewMixConfig("Node2", "localhost", "3332", pub2)
	m3 := config.NewMixConfig("Node3", "localhost", "3333", pub3)

	nodes := []config.MixConfig{m1, m2, m3}

	x := big.NewInt(100)
	asb, err := getSharedSecrets(curve, nodes, *x)
	if err != nil {
		t.Error(err)
	}

	encMsg, err := encapsulateContent(asb, message)
	if err != nil {
		t.Error(err)
	}

	var decMsg []byte

	decMsg = encMsg
	privs := [][]byte{priv1, priv2, priv3}
	for i, v := range privs {
		decMsg, err = ProcessSphinxPayload(asb[i].Alpha, decMsg, v)
		if err != nil {
			t.Error(err)
		}
	}
	assert.Equal(t, []byte(message), decMsg)
}
